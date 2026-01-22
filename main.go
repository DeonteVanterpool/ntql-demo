// web.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DeonteVanterpool/ntql"
)

func parseHandler(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("q")

	// Use the lexer
	lexer := ntql.NewLexer(input)
	tokens, err := lexer.Lex()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Parse
	parser := ntql.NewParser(tokens)
	query, err := parser.Parse()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Generate SQL
	sql, err := query.ToSQL()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"sql": sql})
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func suggestHandler(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("q")

	// Create completion engine
	tags := []string{"school", "work", "projects"}
	engine := ntql.NewCompletionEngine(tags)

	// Get suggestions
	suggestions, err := engine.Suggest(input)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	suggestions = removeDuplicateStr(suggestions)

	json.NewEncoder(w).Encode(map[string]interface{}{"suggestions": suggestions})
}

func main() {
	// Serve simple HTML page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
		<!DOCTYPE html>
		<html>
		<head><title>NTQL Demo</title></head>
		<body>
			<h1>NTQL Demo</h1>
			
			<h2>Parse Query</h2>
			<input id="query" type="text" value="tag.equals(work)" style="width:300px">
			<button onclick="parse()">Parse</button>
			<pre id="sql"></pre>
			
			<h2>Get Suggestions</h2>
			<input id="partial" type="text" value="tag." style="width:300px">
			<button onclick="suggest()">Suggest</button>
			<div id="suggestions"></div>
			
			<script>
				function parse() {
					var query = document.getElementById('query').value;
					fetch('/parse?q=' + encodeURIComponent(query))
						.then(r => r.json())
						.then(data => {
							if (data.error) {
								document.getElementById('sql').textContent = 'Error: ' + data.error;
							} else {
								document.getElementById('sql').textContent = 'SELECT * FROM tasks WHERE (' + data.sql + ')';
							}
						});
				}
				
				function suggest() {
					var partial = document.getElementById('partial').value;
					fetch('/suggest?q=' + encodeURIComponent(partial))
						.then(r => r.json())
						.then(data => {
							var div = document.getElementById('suggestions');
							if (data.error) {
								div.innerHTML = 'Error: ' + data.error;
							} else if (data.suggestions && data.suggestions.length) {
								div.innerHTML = data.suggestions.join(', ');
							} else {
								div.innerHTML = 'No suggestions';
							}
						});
				}
				
				function setExample(example) {
					document.getElementById('query').value = example;
					parse();
				}
				
				// Parse initial example
				parse();
			</script>
		</body>
		</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// API endpoints
	http.HandleFunc("/parse", parseHandler)
	http.HandleFunc("/suggest", suggestHandler)

	fmt.Println("Starting web demo at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
