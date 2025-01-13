package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	// Initialize the database
	var err error
	db, err = sql.Open("sqlite3", "/data/items.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL,
		singleword INTEGER DEFAULT 0,
		count INTEGER DEFAULT 0
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Provide some examples on a new database
	// Check if the table already has data
	row := db.QueryRow("SELECT COUNT(name) FROM items")
	var rowCountItems int
	err = row.Scan(&rowCountItems)
	if err != nil {
		log.Fatalf("Failed to count rows: %v", err)
	}

	if rowCountItems == 0 {
		log.Println("Database has been created")
		log.Println("Importing some examples")
		// Insert some default settings
		insertDataQuery := `
		INSERT INTO items (name, url, singleword, count) VALUES ('b', 'https://www.bbc.com', 0, 0);
		INSERT INTO items (name, url, singleword, count) VALUES ('bb', 'https://www.bbc.com/news/world/%s', 0, 0);
		INSERT INTO items (name, url, singleword, count) VALUES ('bbc', 'https://www.bbc.com/search?q=%s', 0, 0);
		`
		_, err = db.Exec(insertDataQuery)
		if err != nil {
			log.Fatalf("Failed to insert data: %v", err)
		} 
	}
	// Create a table for logging queries
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS queries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		keyword TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Create a table for extra settings
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		setting TEXT NOT NULL UNIQUE,
		value TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	// Check if the table already has data
	row2 := db.QueryRow("SELECT COUNT(setting) FROM settings")
	var rowCount int
	err = row2.Scan(&rowCount)
	if err != nil {
		log.Fatalf("Failed to count rows: %v", err)
	}

	if rowCount == 0 {
		log.Println("Setting Duckduckgo as the fallback search engine")
		// Insert some default settings
		insertDataQuery := `
		INSERT INTO settings (setting, value) VALUES
			('fallback_url', 'https://www.duckduckgo.com/?q={searchTerms}');`
		_, err = db.Exec(insertDataQuery)
		if err != nil {
			log.Fatalf("Failed to insert data: %v", err)
		} 
	} else {
			log.Println("Database exists")
	}

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/opensearch.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/opensearchdescription+xml")
		http.ServeFile(w, r, filepath.Join("./static", "opensearch.xml"))
	})

	// Handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/go/", handleRedirect)
	http.HandleFunc("/reset/", handleReset)
	http.HandleFunc("/reset-all", handleResetAll)
	http.HandleFunc("/mod/", handleMod)
	http.HandleFunc("/mod-post/", handleModPost)
	http.HandleFunc("/fallback/", handleFallback)
	http.HandleFunc("/fallback-post/", handleFallbackPost)
	http.HandleFunc("/del/", handleDel)
	http.HandleFunc("/clear/", handleClear)
	http.HandleFunc("/help/", handleHelp)

	// Start the server
	log.Println("GoMarks 🐇 is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Fetch items from the database
	rows, err := db.Query("SELECT id, name, url, singleword, count FROM items ORDER BY name ASC")
	if err != nil {
		http.Error(w, "Failed to fetch items.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Fetch the last 10 queries from the queries table
	queriesRows, err := db.Query("SELECT keyword, created_at FROM queries ORDER BY created_at DESC LIMIT 30")
	if err != nil {
		http.Error(w, "Failed to fetch queries.", http.StatusInternalServerError)
		return
	}
	defer queriesRows.Close()

	var items []struct {
		ID    int
		Name  string
		URL   string
		Singleword int
		Count int
	}
	for rows.Next() {
		var item struct {
			ID    int
			Name  string
			URL   string
			Singleword int
			Count int
		}
		if err := rows.Scan(&item.ID, &item.Name, &item.URL, &item.Singleword, &item.Count); err != nil {
			http.Error(w, "Failed to parse items.", http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	var queries []struct {
		Keyword   string
		CreatedAt string
	}
	for queriesRows.Next() {
		var query struct {
			Keyword   string
			CreatedAt string
		}
		if err := queriesRows.Scan(&query.Keyword, &query.CreatedAt); err != nil {
			http.Error(w, "Failed to parse queries.", http.StatusInternalServerError)
			return
		}
		queries = append(queries, query)
	}

	// Get current fallback engine
	var fallback_url string
	err = db.QueryRow("SELECT value FROM settings WHERE setting='fallback_url'").Scan(&fallback_url)
	if err != nil {
		http.Error(w, "Fallback URL not found.", http.StatusNotFound)
		return
	}

	// Count number of items
	var countlinks string
	err = db.QueryRow("SELECT COUNT(name) FROM items").Scan(&countlinks)
	if err != nil {
		http.Error(w, "Failed to count items.", http.StatusInternalServerError)
		return
	}

	// Render the index page
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
	    <link rel="search" type="application/opensearchdescription+xml" href="/opensearch.xml" title="GoMarks">
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>GoMarks</title>
		<link rel="stylesheet" href="/static/style.css">
		<link rel="icon" type="image/png" sizes="32x32" href="/static/favicon.png">
		<script>
			let sortOrder = {
				keyword: 'asc',
				url: 'asc',
				visits: 'asc'
			};

			function sortTable(columnIndex) {
				const table = document.querySelector('table');
				const rows = Array.from(table.querySelectorAll('tr:nth-child(n+2)'));
				const columnType = table.rows[0].cells[columnIndex].dataset.type;
				
				rows.sort((rowA, rowB) => {
					let cellA = rowA.cells[columnIndex].innerText.trim();
					let cellB = rowB.cells[columnIndex].innerText.trim();

					if (columnType === 'numeric') {
						cellA = parseInt(cellA, 10);
						cellB = parseInt(cellB, 10);
					}

					if (sortOrder[table.rows[0].cells[columnIndex].dataset.sort] === 'asc') {
						if (cellA < cellB) return -1;
						if (cellA > cellB) return 1;
					} else {
						if (cellA < cellB) return 1;
						if (cellA > cellB) return -1;
					}
					return 0;
				});

				rows.forEach(row => table.appendChild(row)); // Reattach sorted rows

				// Toggle sorting order
				sortOrder[table.rows[0].cells[columnIndex].dataset.sort] = sortOrder[table.rows[0].cells[columnIndex].dataset.sort] === 'asc' ? 'desc' : 'asc';
			}
		</script>
	</head>
	<body>
		<h2><a href=".">GoMarks <img src="/static/favicon.png" width="32" height="32"></a></h2>

		<a href="/help">Help</a> | <a href="https://github.com/sebw/GoMarks/">v0.1</a> | 👨‍💻 <a href="https://github.com/sebw/">@sebw</a>

		<p><form action="/go/" method="get" target="_blank">
			<input type="text" name="q" placeholder="Search here or make GoMarks your default search engine " required>
			<button type="submit">Search</button>
		</form></p>

		<h2>You have {{.Countlinks}} shortcuts</h2>

		<p><form action="/add" method="post">
			<input type="text" name="name" placeholder="Keyword" required>
			<input type="url" name="url" id="url" placeholder="Destination URL" size="100" required autocomplete="off">
			<label for="singleword">Force 1️⃣ single word placeholder</label>
			<input type="checkbox" id="singleword" name="singleword" disabled> (<a href="/help/#placeholder">?</a>)
			<button type="submit">Add new shortcut</button>
		</form></p>
		<script>
		// enable checkbox if placeholder in URL
		const textField = document.getElementById('url');
		const checkbox = document.getElementById('singleword');
	
		// Add an input event listener to the text field
		url.addEventListener('input', () => {
		  // Enable the checkbox if the URL contains "%s" (case-insensitive)
		  const value = textField.value	;
		  singleword.disabled = !value.includes('%s');
		});
		</script>

		<p><a href="/fallback">Fallback search engine</a> <code>{{.Fallback}}</code></p>

		<table class="links">
			<tr>
				<th data-sort="keyword" data-type="string" onclick="sortTable(0)">Keyword ↕️</th>
				<th style="width: 200px;" data-sort="url" data-type="string" onclick="sortTable(1)">Destination URL ↕️</th>
				<th style="text-align: center;" data-sort="visits" data-type="numeric" onclick="sortTable(2)">Visits count ↕️</th>
				<th style="text-align: center;">Reset visits count</th>
				<th style="text-align: center;">Edit</th>
				<th style="text-align: center;">Delete</th>
			</tr>
			{{range .Items}}
			<tr>
				<td><code id="keyword"><a href="/go/?q={{.Name}}" target="_blank">{{.Name}}</code> {{if eq .Singleword 1}} 1️⃣{{end}}</a></td>
				<td><a href="{{.URL}}" target="_blank">{{.URL}}</a></td>
				<td style="text-align: center;">{{.Count}}</td>
				<td style="text-align: center;"><a href="/reset/{{.Name}}">♻️</a></td>
				<td style="text-align: center;"><a href="/mod/{{.Name}}">✍🏻</a></td>
				<td style="text-align: center;"><a href="/del/{{.Name}}">❌</a></td>
			</tr>
			{{end}}
		</table>

		<h2>Last 30 queries</h2>

		<table class="history">
			<tr>
				<th>Query</th>
				<th>Time</th>
			</tr>
			{{range .Queries}}
			<tr>
				<td><a href="/go/?q={{.Keyword}}" target="_blank">{{.Keyword}}</td>
				<td>{{.CreatedAt}}</td>
			</tr>
			{{end}}
		</table>

		<h2 id="admin">Administration</h2>

		<p><a href="/reset-all">Reset all visit counts</a></p>
		<a id="deleteLink" href="#" onclick="confirmDelete(event)">Delete Full Queries History</a>
		
		<div id="confirmDelete" style="display:none;">
			Are you sure you want to delete? 
			<button onclick="proceedDelete()">Yes</button>
			<button onclick="cancelDelete()">No</button>
		</div>
		
		<script>
			function confirmDelete(event) {
				event.preventDefault(); // Prevent the link from executing its default action
				document.getElementById('confirmDelete').style.display = 'inline';
				document.getElementById('deleteLink').style.display = 'none';
			}
		
			function proceedDelete() {
				window.location.href = '/clear';
				document.getElementById('confirmDelete').style.display = 'none';
			}
		
			function cancelDelete() {
				document.getElementById('confirmDelete').style.display = 'none';
				document.getElementById('deleteLink').style.display = 'inline';
			}
		</script>

		<p><a href="/fallback">Configure fallback search engine</a></p>



	</body>
	</html>
	`

	tmplParsed := template.Must(template.New("index").Parse(tmpl))
	tmplParsed.Execute(w, struct {
		Items   []struct {
			ID    int
			Name  string
			URL   string
			Singleword int
			Count int
		}
		Queries []struct {
			Keyword   string
			CreatedAt string
		}
		Fallback string
		Countlinks string
	}{
		Items:   items,
		Queries: queries,
		Fallback: fallback_url,
		Countlinks: countlinks,
	})
}



func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method.", http.StatusMethodNotAllowed)
		return
	}

	name := r.FormValue("name")
	url := r.FormValue("url")
	singlewordvalue := r.FormValue("singleword")
	if name == "" || url == "" {
		http.Error(w, "Keyword and URL cannot be empty.", http.StatusBadRequest)
		return
	}

	var singleword int
	// singleword
	if singlewordvalue == "" {
		singleword = 0
	} 
	
	if singlewordvalue == "on" {
		singleword = 1
	}

	// Max one placeholder in URL
	placeholder_count := strings.Count(url, "%s")

	if placeholder_count > 1 {
		http.Error(w, "You can only have one placeholder in your URL.", http.StatusMethodNotAllowed)
		return
	} 

	_, err := db.Exec("INSERT INTO items (name, url, singleword) VALUES (?, ?, ?)", name, url, singleword)
	if err != nil {
		http.Error(w, "Failed to add shortlink. Ensure the keyword is unique.", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	query := queryValues.Get("q")

	// Making vars available in the whole function
	var url string
	var fallback_url string
	var destination_url string
	var singleword int
	var placeholder_present bool
	var words_counting int
	var keyword string
	var second_word string
	var second_word_and_all string

	// Fail if no query provided
	if query == "" {
		http.Error(w, "A query can't be empty.", http.StatusBadRequest)
		return
	}

	// Getting the fallback URL as it will come handy
	err := db.QueryRow("SELECT value FROM settings WHERE setting='fallback_url'").Scan(&fallback_url)
	if err != nil {
		http.Error(w, "Fallback URL not found.", http.StatusInternalServerError)
		return
	}

	// Query has been provided, let's get to work
	if query != "" {
		// Add query to history
		_, err := db.Exec("INSERT INTO queries (keyword) VALUES (?)", query)
		if err != nil {
			http.Error(w, "Failed to log query.", http.StatusInternalServerError)
			return
		}	

		// Manipulate the query
		words := strings.Fields(query)
		words_counting = len(words)

		if words_counting == 1 {
			keyword = words[0]
		}
		if words_counting == 2 {
			keyword = words[0]
			second_word = words[1]
		}
		if words_counting >= 2 {
			keyword = words[0]
			second_word = words[1]
			second_word_and_all = strings.Join(words[1:], " ")
		}
		

		// Assess the first word in the query and check if a keyword matches
		// We force lowercase to make things case insensitive
		var keyword_found int
		err = db.QueryRow("SELECT COUNT(name) FROM items where LOWER(name) = LOWER(?)", keyword).Scan(&keyword_found)
		if err != nil {
			http.Error(w, "Failed to count.", http.StatusInternalServerError)
			return
		}

		// Scenario
		// Keyword not found
		// Outcome: pass the full query to the fallback URL
		if keyword_found == 0 {
			url = strings.ReplaceAll(fallback_url, "{searchTerms}", query)
		}

		// if keyword found, fetch the destination URL
		if keyword_found == 1 {
			err = db.QueryRow("SELECT url, singleword FROM items WHERE LOWER(name) = LOWER(?)", keyword).Scan(&destination_url, &singleword)
			if err != nil {
				http.Error(w, "Failed to retrieve destination URL.", http.StatusInternalServerError)
				return
			}
		}

		// Check if URL contains a placeholder. In this case, we expect at least two words
		if strings.Contains(destination_url, "%s") {
			placeholder_present = true
		} else {
			placeholder_present = false
		}

		// Scenarios
		// a single keyword and the URL doesn't have a placeholder
		// Outcome: redirection to the URL
		if keyword_found == 1 && words_counting == 1 && !placeholder_present {
			url = destination_url
		}

		// two or more words are present but there's no placeholder
		// outcome: failure
		if keyword_found == 1 && words_counting > 1 && !placeholder_present {
			http.Error(w, "The keyword " + keyword + " doesn't accept options as its URL " + destination_url + " doesn't have a placeholder.", http.StatusBadRequest)
			return
		}

		// a single keyword but the URL has a placeholder
		// Outcome: failure, a second word is expected
		if keyword_found == 1 && words_counting == 1 && placeholder_present {
			http.Error(w, "The keyword " + keyword + " expects an option as its URL " + destination_url + " contains a placeholder.", http.StatusBadRequest)
			return
		}

		// option(s) are passed, a placeholder is present and single word is enforced
		// no failure is expected here
		if keyword_found == 1 && words_counting >= 2 && placeholder_present && singleword == 1 {
			// matching the expected number of options
			// outcome: success, taking to destination (ex: docker alpine)
			if words_counting == 2 {
				url = strings.ReplaceAll(destination_url, "%s", second_word)
			}
			// more than one word specified while single word is expected
			// outcome: not taking to destination URL but fallback (ex: docker versus kubernetes)
			if words_counting > 2 {
				url = strings.ReplaceAll(fallback_url, "{searchTerms}", query)
			}
		}

		// two or more words, URL expects an option but it can be multiple words (ex: amazon search)
		if keyword_found == 1 && words_counting >= 2 && placeholder_present && singleword == 0 {
			url = strings.ReplaceAll(destination_url, "%s", second_word_and_all)
		}

		// update the visit count
		db.Exec("UPDATE items SET count = count + 1 WHERE name = ?", keyword)
		
		// Final call
		http.Redirect(w, r, url, http.StatusFound)
		}
}
func handleReset(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/reset/"):]
	if name == "" {
		http.Error(w, "Keyword is required to reset the visit counter.", http.StatusBadRequest)
		return
	}

	db.Exec("UPDATE items SET count = 0 WHERE name = ?", name)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleResetAll(w http.ResponseWriter, r *http.Request) {
	db.Exec("UPDATE items SET count = 0")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleMod(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/mod/"):]
	if name == "" {
		http.Error(w, "Keyword is required to modify a link.", http.StatusBadRequest)
		return
	}

	var item struct {
		ID   int
		Name string
		URL  string
		Singleword int
		Checkbox string
	}


	err := db.QueryRow("SELECT id, name, url, singleword FROM items WHERE LOWER(name) = LOWER(?)", name).Scan(&item.ID, &item.Name, &item.URL, &item.Singleword)
	if err != nil {
		http.Error(w, "Keyword not found.", http.StatusNotFound)
		return
	}

	// defining the state of the checkbox
	if strings.Contains(item.URL, "%s") {
		item.Checkbox = "enabled"
	} else {
		item.Checkbox = "disabled"
	}

	// Render the edit page
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>GoMarks - Edit Link</title>
		<link rel="stylesheet" href="/static/style.css">
		<script>
		function goToIndex() {
			window.location.href = "/";
		}
		</script>
	</head>
	<body>
		<h2><a href="/">Edit Link</a></h2>
		<form action="/mod-post/{{.Name}}" method="post">
			<input type="text" name="name" value="{{.Name}}" placeholder="Keyword"required>
			<input type="url" name="url" id="url" value="{{.URL}}" placeholder="Destination URL" required autocomplete="off">
			<label for="singleword">Force 1️⃣ single word placeholder</label>
			<input type="checkbox" id="singleword" name="singleword" {{if eq .Singleword 1}}checked{{end}} {{.Checkbox}}> (<a href="/help/#placeholder">?</a>)
			<button type="submit">Save</button></p>
			<button type="button" onclick="goToIndex()">Cancel</button>
		</form>
		<script>
		// enable checkbox if placeholder in URL
		const textField = document.getElementById('url');
		const checkbox = document.getElementById('singleword');
		
		// Add an input event listener to the text field
		url.addEventListener('input', () => {
		  // Enable the checkbox if the URL contains "%s"
		  const value = textField.value	;
		  singleword.disabled = !value.includes('%s');
		});
		</script>

	</body>
	</html>
	`

	tmplParsed := template.Must(template.New("edit").Parse(tmpl))
	tmplParsed.Execute(w, item)
}

func handleModPost(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/mod-post/"):]
	if name == "" {
		http.Error(w, "Keyword is required to modify a link.", http.StatusBadRequest)
		return
	}

	// Get updated name and URL from the form
	newName := r.FormValue("name")
	url := r.FormValue("url")
	singleword := r.FormValue("singleword")

	var singlewordvalue int

	if singleword == "on" {
		singlewordvalue = 1
	} else {
		singlewordvalue = 0
	}

	if newName == "" || url == "" {
		http.Error(w, "Keyword and URL cannot be empty.", http.StatusBadRequest)
		return
	}

	// Update the link in the database
	_, err := db.Exec("UPDATE items SET name = ?, url = ?, singleword = ? WHERE name = ?", newName, url, singlewordvalue, name)
	if err != nil {
		http.Error(w, "Failed to update the link.", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleDel(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/del/"):]
	if name == "" {
		http.Error(w, "Keyword is required to delete a link.", http.StatusBadRequest)
		return
	}

	// Delete the shortlink from the database
	_, err := db.Exec("DELETE FROM items WHERE name = ?", name)
	if err != nil {
		http.Error(w, "Failed to delete link.", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleFallback(w http.ResponseWriter, r *http.Request) {
	var item struct {
		Value string
	}
	err := db.QueryRow("SELECT value FROM settings WHERE setting='fallback_url'").Scan(&item.Value)
	if err != nil {
		http.Error(w, "Fallback URL not found.", http.StatusNotFound)
		return
	}

	// Render the edit page
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>GoMarks</title>
		<link rel="stylesheet" href="/static/style.css">
	    <script>
        function goToIndex() {
            window.location.href = "/";
        }
    </script>
	</head>
	<body>
		<h2><a href="/">Configure fallback search engine</a></h2>
		<form action="/fallback-post/" method="post">
			<input type="url" name="url" value="{{.Value}}" required autocomplete="off">
			<button type="submit">Save</button></p>
			<button type="button" onclick="goToIndex()">Cancel</button>
		</form>
		⚠️ <code>{searchTerms}</code> is required in the URL.
		<p>
		Here's a list of popular search engines to choose from:</p>
		<table class="fallback">
        <thead>
            <tr>
                <th>Search engine</th>
                <th>Search URL</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>Google</td>
                <td><code>https://www.google.com/search?q={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Bing</td>
                <td><code>https://www.bing.com/search?q={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Yahoo</td>
                <td><code>https://search.yahoo.com/search?p={searchTerms}</code></td>
            </tr>
            <tr>
                <td>DuckDuckGo</td>
                <td><code>https://www.duckduckgo.com/?q={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Baidu</td>
                <td><code>https://www.baidu.com/s?wd={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Yandex</td>
                <td><code>https://yandex.com/search/?text={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Ecosia</td>
                <td><code>https://www.ecosia.org/search?q={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Ask</td>
                <td><code>https://www.ask.com/web?q={searchTerms}</code></td>
            </tr>
            <tr>
                <td>AOL Search</td>
                <td><code>https://search.aol.com/aol/search?q={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Startpage</td>
                <td><code>https://www.startpage.com/do/dsearch?query={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Qwant</td>
                <td><code>https://www.qwant.com/?q={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Swisscows</td>
                <td><code>https://swisscows.com/web?query={searchTerms}</code></td>
            </tr>
            <tr>
                <td>Brave Search</td>
                <td><code>https://search.brave.com/search?q={searchTerms}</code></td>
            </tr>
        </tbody>
    </table>
	</body>
	</html>
	`

	tmplParsed := template.Must(template.New("edit").Parse(tmpl))
	tmplParsed.Execute(w, item)
}

func handleFallbackPost(w http.ResponseWriter, r *http.Request) {
	// Get updated fallback URL from the form
	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "Fallback URL cannot be empty.", http.StatusBadRequest)
		return
	} 

	// Max one placeholder in URL
	placeholder_count := strings.Count(url, "{searchTerms}")

	if placeholder_count != 1 {
		http.Error(w, "You need exactly one {searchTerms} in your fallback URL.", http.StatusMethodNotAllowed)
		return
	} 

	// Update the fallback URL in the database
	_, err := db.Exec("UPDATE settings SET value = ? WHERE setting='fallback_url'", url)
	if err != nil {
		http.Error(w, "Failed to update fallback URL.", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleClear(w http.ResponseWriter, r *http.Request) {
	db.Exec("DELETE FROM queries WHERE 1 == 1")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleHelp(w http.ResponseWriter, r *http.Request) {
	// Define the HTML content as a template
	tmpl := template.Must(template.New("static").Parse(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>GoMarks Help</title>
		<link rel="stylesheet" href="/static/style.css">
	</head>
	<body>

	<h2><a href="/">GoMarks Help</a></h2>

	<a href="/help/#simple">Simple shortcuts</a> | <a href="/help/#smart">Smart shortcuts</a> | <a href="/help/#smarter">Smarter shortcuts</a>
	<br>
	<a href="/help/#browser">GoMarks as your search engine</a> | <a href="/help/#mobile">GoMarks for iPhone</a> | <a href="/help/#fallback">Fallback search engine</a>
<br><br><br>
	GoMarks let's you create shortcuts (similar to <a href="https://meta.wikimedia.org/wiki/Go_links" target="_blank">go links</a>, <a href="https://en.wikipedia.org/wiki/Smart_bookmark" target="_blank">smart bookmarks</a> or <a href="https://en.wikipedia.org/wiki/DuckDuckGo#Bangs" target="_blank"/>bangs</a>) that redirect to websites.</p>

	Shortcuts require two things: a keyword and a destination URL.</p>

	To use your shortcuts, you can use the GoMarks webpage, call the URL <code>https://gomarks.example.com/go/?q=your query</code>, or make GoMarks your browser's <a href="/help/#browser">default search engine</a>.</p>
	
	On iPhone, you can create a <a href="/help/#mobile">simple automation</a> that take advantage of GoMarks shortcuts.</p>
	
	Making GoMarks your default search engine is the fastest way to consume your shortcuts (and impress your colleagues!).</p>
	<br>
	<h3 id="simple">Simple shortcuts</h3>


	A keyword <code>bbc</code> can take you to the destination URL <code>https://www.bbc.com</code>.</p>

	Simple shortcuts can't take options (like <code>bbc europe</code>) because they are simple, not smart.
	
	<br>
	<h3 id="smart">Smart shortcuts</h3>

	Smart shortcuts use a placeholder <code>%s</code> in destination URLs.</p>

	With this destination URL <code>https://www.bbc.com/news/world/%s</code>, I can now search for <code>bbc europe</code> or <code>bbc australia</code></p>

	Smart shortcuts require an option. You can no longer use <code>bbc</code> alone.

	<h3 id="smarter">Smarter shortcuts</h3>

	We can improve <code>bbc</code> further by searching in BBC's articles.</p>

    Let's search the terms "open source":</p>
	
	Actual URL <code>https://www.bbc.com/search?q=<span style="background-color:#bf616a;">open+source</span>&edgeauth=eyJhbGciOi...</code> ➡️ Destination URL <code>https://www.bbc.com/search?q=<span style="background-color:#bf616a;">%s</span></code></p>
	
	⚠️ Queries are traditionally passed behind <code>q=</code>, <code>query=</code> or <code>search=</code> arguments but nothing prevents a website owner to use <code>banana=</code>.</p>
	
	You can sometimes remove a lot of garbage from the URL, like <code>&edgeauth=eyJhbGciOi</code> in the example above.</p>

	Searching for <code>bbc <span style="background-color:#bf616a;">open source</span></code> would take you to a list of BBC articles about open source.</p>

	Some websites do not pass searches in arguments but rather directly in the URL.</p>
	
	See the <code>https://http.cat/<span style="background-color:#bf616a;">%s</span></code> example below.

	<h4 id="placeholder">Single or Multi Words Placeholder</h4>

	When adding or editing a shortcut, there's a "Force 1️⃣ single word placeholder" option.</p>

	This option works with placeholders, the checkbox is only active if your destination URL contains <code>%s</code>.</p>

	This option helps GoMarks choose the best solution between the destination URL or the <a href="/help/#fallback">fallback search engine</a>.</p>

	To illustrate the concept, let's take queries around the Docker topic.</p>

	You want to be able to find Docker images in Docker Hub by using <code>docker mariadb</code> (see destination URL in examples below).</p>

	You are also likely to search <code>docker versus openshift</code> or <code>docker compose syntax</code>.</p>

	In the latter scenario, without single word placeholder, you would be taken to Docker Hub with a search for <code>versus openshift</code>. This would provide very odd results.</p>

	When single word is enabled, an icon 1️⃣ appears next to the keyword.</p>

	<h4>Examples</h4>

	<table class="links">
	<tr>
		<th>Description</th>
		<th>Link</th>
	</tr>
	<tr>
		<td>Find Docker images</td>
		<td><code>https://hub.docker.com/search?q</span>=<span style="background-color:#bf616a;">%s</span></code></td>
	</tr>
	<tr>
		<td>Search in author's Github repositories</td>
		<td><code>https://github.com/sebw?tab=repositories&q</span>=<span style="background-color:#bf616a;">%s</span>&type=&language=&sort=</code></td>
	</tr>
	<tr>
		<td>Search a Fedora Linux package</td>
		<td><code><code>https://packages.fedoraproject.org/search?query</span>=<span style="background-color:#bf616a;">%s</span></code></code></td>
	</tr>
	<tr>
		<td>Compare Amazon prices</td>
		<td><code>https://keepa.com/#!search/4-<span style="background-color:#bf616a;">%s</span></code></td>
	</tr>
	<tr>
		<td>Find a Python Library</td>
		<td><code>https://pypi.org/search/?q</span>=<span style="background-color:#bf616a;">%s</span></code></td>
	</tr>
	<tr>
		<td>Wikipedia Search</td>
		<td><code>https://en.wikipedia.org/?search</span>=<span style="background-color:#bf616a;">%s</span></code></td>
	</tr>
	<tr>
		<td>Eggtimer. Query "timer 5m" to set a 5 minute timer.</td>
		<td><code>https://e.ggtimer.com/<span style="background-color:#bf616a;">%s</span></code></td>
	</tr>
	<tr>
		<td>Unsure about your pronounciation? Query "say miscellaneous"</td>
		<td><code>https://www.google.com/search?q</span>=pronounce+<span style="background-color:#bf616a;">%s</span></code></td>
	</tr>
	<tr>
		<td>Check movie rating</td>
		<td><code>https://www.imdb.com/find?q=<span style="background-color:#bf616a;">%s</span></code></td>
	</tr>
	<tr>
		<td>Check HTTP status code for cat lovers</td>
		<td><code>https://http.cat/<span style="background-color:#bf616a;">%s</span></code></td>
	</tr>
	</table>
<br>
	<h3 id="browser">Making GoMarks Your Default Search Engine</h3>

	By making GoMarks your default search engine, you can type your queries in the URL/search bar for faster access to your links.</p>

	GoMarks won't leave you hanging if you make a request that doesn't match any keyword, it would just redirect to the <a href="/help/#fallback">fallback search engine</a>.</p>

	<h4>Chrome</h4>

	In the address bar go to <code>chrome://settings/searchEngines</code></p>

	Search Engine > Manage search engines and site search</p>

	Site Search > Add</p>

	Add Site Search and use this URL <code>https://yourGoMarks/go/%s</code></p>

	<img src="/static/help/chrome_step1.png"></p>

	Click on the hamburger menu for your new search engine > Make default</p>

	<img src="/static/help/chrome_step2.png"></p>

	Start searching</p>

	<img src="/static/help/chrome_step3.png"></p>

	<h4>Firefox</h4>

	Go to your GoMarks instance > Right click on the URL > Add "GoMarks"</p>

	<img src="/static/help/firefox_step1.png"></p>

	Go in Firefox settings > Change Default Search Engine</p>

	<img src="/static/help/firefox_step2.png"></p>

	Start searching</p>

	<img src="/static/help/firefox_step3.png"></p>

	<h3>Firefox - Alternative Way</h3>

	If the method above doesn't work, use <a href="https://addons.mozilla.org/en-GB/firefox/addon/add-custom-search-engine/">this add-on</a>.</p>
<br>
	<h3 id="mobile">iPhone Shortcuts</h3>

	You can't change the default search engine in Safari on iPhone.</p>

	You can use an iOS shortcut and use the action button or widgets to call GoMarks.</p>

	<img height="600" src="/static/help/iphone_step1.jpg">
	<img height="600" src="/static/help/iphone_step2.jpg"></p>
<br>
	<h2 id="fallback">Fallback Search Engine</h3>

	If your request doesn't match any shortcut or if you use <a href="/help/#placeholder">single word placeholders</a>, GoMarks can sends your request to the fallback search engine.</p>

	The fallback search engine is <a href="/fallback">configurable</a> (Google, Duckduckgo, your own self-hosted solution, etc.)</p>
	<br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br><br>
	🍌🐇<br><br>
	</body>
	</html>
	`))

	// Render the template
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

