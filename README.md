# LiteDB-Go
LiteDB-Go is a lightweight, file-based database engine written in Go.  
It stores each record as a JSON document on disk and provides a simple, clean API for reading, writing, listing, and deleting structured data.  
The project is designed to help understand low-level storage concepts, concurrency control, and atomic writes without relying on external databases.

---

## ✨ Features

- **JSON Document Storage**: Each record is saved as its own `.json` file.
- **Atomic Writes**: Uses temporary files and safe renaming to avoid partial data.
- **Collection-Based Structure**: Groups related records inside dedicated folders.
- **Concurrency-Safe Design**: Per-collection mutexes ensure safe parallel access.
- **Simple API**: Easy-to-use `Write`, `Read`, `ReadAll`, and `Delete` functions.
- **Optional Logging**: Integrates with `lumber` for configurable debug output.

---

## Initialize the project
```bash
go mod init github.com/SagarDas211/LiteDB-Go
go get github.com/jcelliott/lumber
```

## Run the Project
```bash
go run main.go
```

## 🧩 Usage Example
```bash
db, err := New("./data", nil)
if err != nil {
    panic(err)
}

user := User{
    Name: "John",
    Age:  "30",
}

db.Write("users", user.Name, user)

records, _ := db.ReadAll("users")
fmt.Println(records)
```

## ⚠️ Current Limitations
- No support for updating existing records
- No query or filtering mechanism
- No indexing for faster lookups
- No transactional guarantees
- Designed for learning purposes, not production use

## 🔮 Future Improvements
- Add update and partial update functionality
- Implement query and filtering support
- Introduce indexing for optimized reads
- Add unit and integration test coverage
- Provide REST API and CLI interfaces