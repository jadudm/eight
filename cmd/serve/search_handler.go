package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/sqlite/schemas"
	"github.com/jadudm/eight/internal/util"
	"github.com/kljensen/snowball"
	"go.uber.org/zap"
)

// FIXME This becomes the API search interface
type ServeRequestInput struct {
	Host  string `json:"host"`
	Terms string `json:"terms"`
}

var statmap sync.Map

func permutate[T any](data []T) [][]T {
	if len(data) == 0 {
		return nil
	}

	permutation := make([]T, len(data))
	indexInUse := make([]bool, len(data))

	var ret [][]T
	var f func(idx int)

	f = func(idx int) {
		if idx >= len(data) {
			arr := make([]T, len(data))
			copy(arr, permutation)
			ret = append(ret, arr)
			return
		}
		for i := 0; i < len(data); i++ {
			if !indexInUse[i] {
				indexInUse[i] = true
				permutation[idx] = data[i]
				f(idx + 1)
				indexInUse[i] = false
			}
		}
	}

	f(0)
	return ret
}

// Too few results? Permute!
// 1. Take all the search terms (if there are more than two)
// 2. Permute
// 3. Delete the last term in each permutation
// 4. Run those queries
// 5. Keep the top `n` of each
// 6. Interleave, and return

func permuteSubqueries(queries *schemas.Queries,
	improved_terms []string,
	results_per_query int64) []schemas.SearchSiteIndexSnippetsRow {
	permuted := permutate(improved_terms)
	zap.L().Info("permutated", zap.String("permuted", fmt.Sprintln(permuted)))

	shorter_queries := util.Map(permuted,
		func(item []string) []string {
			return item[0 : len(item)-1]
		})

	// zap.L().Info("shorter queries", zap.String("shorts", fmt.Sprintln(permuted)))

	combined := make([][]schemas.SearchSiteIndexSnippetsRow, 0)
	for _, q := range shorter_queries {
		res2, _ := queries.SearchSiteIndexSnippets(context.Background(), schemas.SearchSiteIndexSnippetsParams{
			Text:  strings.Join(q, " "),
			Limit: results_per_query,
		})
		combined = append(combined, res2)
	}

	// zap.L().Info("combined results",
	// 	zap.String("combos", fmt.Sprintln(permuted)),
	// 	zap.String("result_combos", fmt.Sprintln(combined)),
	// 	zap.Int("combo length", len(combined[0])),
	// )

	interleaved := make([]schemas.SearchSiteIndexSnippetsRow, 0)
	max_result_set_length := 0
	for _, set := range combined {
		if len(set) > max_result_set_length {
			max_result_set_length = len(set)
		}
	}

	for i := range max_result_set_length {
		for _, c := range combined {
			if i < len(c) {
				interleaved = append(interleaved, c[i])
			}
		}
	}

	return interleaved
}

func runQuery(c *gin.Context, sri ServeRequestInput, limit int) (
	[]schemas.SearchSiteIndexSnippetsRow,
	time.Duration,
	error) {
	start := time.Now()

	s, _ := env.Env.GetUserService("serve")
	database_files_path := s.GetParamString("database_files_path")
	results_per_query := s.GetParamInt64("results_per_query")

	sqlite_file := database_files_path + "/" + sri.Host + ".sqlite"
	if _, err := os.Stat(sqlite_file); errors.Is(err, os.ErrNotExist) {
		duration := time.Since(start)
		c.IndentedJSON(http.StatusOK, gin.H{
			"result":  "err",
			"elapsed": duration,
			"results": nil,
		})
		return nil, time.Since(start), fmt.Errorf("search host not found")
	}

	db, err := sql.Open("sqlite3", sqlite_file)
	if err != nil {
		log.Fatal("SERVCE cannot open SQLite file", sqlite_file)
	}

	// Stem the terms and add wildcards.
	improved_terms := make([]string, 0)
	for _, t := range strings.Split(sri.Terms, " ") {
		stemmed, err := snowball.Stem(t, "english", true)
		if err != nil {
			// Pass. Keep the value as-is
			improved_terms = append(improved_terms, t)
		} else {
			if len(stemmed) > 0 {
				improved_terms = append(improved_terms, stemmed+"*")

			}
		}
	}
	improved_terms_string := strings.Join(improved_terms, " ")

	zap.L().Debug("search string",
		zap.String("original", sri.Terms),
		zap.String("improved", improved_terms_string))

	queries := schemas.New(db)
	res, err := queries.SearchSiteIndexSnippets(context.Background(), schemas.SearchSiteIndexSnippetsParams{
		Text:  improved_terms_string, //sri.Terms,
		Limit: results_per_query,
	})

	// This is all fine and good, but it would be nice to annotate
	// each search result with the terms that were used
	//zap.L().Info("search results", zap.Int("count", len(res)))

	if (len(res) < 3) && (len(improved_terms) > 2) {
		res = permuteSubqueries(queries, improved_terms, results_per_query)
	}

	duration := time.Since(start)
	return res[0:limit], duration, err
}

// //////////////////////////
// Search Handler
// Handles the API requests as they come in.
func SearchHandler(c *gin.Context) {
	var sri ServeRequestInput
	if err := c.BindJSON(&sri); err != nil {
		return
	}

	// These are in another file, because this one should stay shorter.
	// FIXME: the limit of 10 should not be hard-coded
	rows, duration, err := runQuery(c, sri, 10)
	runStats(sri, duration)

	if err != nil {
		c.IndentedJSON(http.StatusOK, gin.H{
			"result":  "err",
			"elapsed": duration,
			"results": nil,
		})
		return
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{
			"result":  "ok",
			"elapsed": duration,
			"results": rows,
		})
		return
	}
}
