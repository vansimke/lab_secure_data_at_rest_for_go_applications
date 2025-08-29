package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type User struct {
	Username string `json:"username"`
	// Insecure: plaintext password for initial state
	Password string `json:"password"`
}

type Note struct {
	Owner string `json:"owner"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Store struct {
	Users []User `json:"users"`
	Notes []Note `json:"notes"`
}

func loadStore(base string) (*Store, error) {
	b, err := os.ReadFile(filepath.Join(base, "store.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Store{}, nil
		}
		return nil, err
	}
	var s Store
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func saveStore(base string, s *Store) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(base, "store.json"), b, 0o600)
}

func register(base, username, password string) error {
	s, err := loadStore(base)
	if err != nil {
		return err
	}
	for _, u := range s.Users {
		if u.Username == username {
			return fmt.Errorf("user exists")
		}
	}
	s.Users = append(s.Users, User{Username: username, Password: password})
	return saveStore(base, s)
}

func login(base, username, password string) error {
	s, err := loadStore(base)
	if err != nil {
		return err
	}
	for _, u := range s.Users {
		if u.Username == username {
			if u.Password == password {
				return nil
			}
			return fmt.Errorf("invalid password")
		}
	}
	return fmt.Errorf("user not found")
}

func addNote(base, owner, title, body string) error {
	s, err := loadStore(base)
	if err != nil {
		return err
	}
	s.Notes = append(s.Notes, Note{Owner: owner, Title: title, Body: body})
	return saveStore(base, s)
}

func listNotes(base, owner string) ([]Note, error) {
	s, err := loadStore(base)
	if err != nil {
		return nil, err
	}
	var out []Note
	for _, n := range s.Notes {
		if n.Owner == owner {
			out = append(out, n)
		}
	}
	return out, nil
}

func main() {
	var (
		dataDir = flag.String("data", "./data", "data directory")
		cmd     = flag.String("cmd", "help", "command: register|login|add-note|list-notes|help")
		user    = flag.String("user", "", "username")
		pass    = flag.String("pass", "", "password")
		title   = flag.String("title", "", "note title")
		body    = flag.String("body", "", "note body")
	)
	flag.Parse()

	if _, err := os.Stat(*dataDir); errors.Is(err, os.ErrNotExist) {
		_ = os.MkdirAll(*dataDir, 0o700)
	}

	switch *cmd {
	case "register":
		if *user == "" || *pass == "" {
			fmt.Println("-user and -pass required")
			os.Exit(2)
		}
		if err := register(*dataDir, *user, *pass); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("registered")
	case "login":
		if *user == "" || *pass == "" {
			fmt.Println("-user and -pass required")
			os.Exit(2)
		}
		if err := login(*dataDir, *user, *pass); err != nil {
			fmt.Println("login failed:", err)
			os.Exit(1)
		}
		fmt.Println("login ok")
	case "add-note":
		if *user == "" || *title == "" {
			fmt.Println("-user and -title required")
			os.Exit(2)
		}
		if err := addNote(*dataDir, *user, *title, *body); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		fmt.Println("note added")
	case "list-notes":
		if *user == "" {
			fmt.Println("-user required")
			os.Exit(2)
		}
		notes, err := listNotes(*dataDir, *user)
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		for _, n := range notes {
			fmt.Printf("- %s: %s\n", n.Title, n.Body)
		}
	case "help":
		fmt.Println("Usage:")
		fmt.Println("  -cmd register -user alice -pass s3cr3t")
		fmt.Println("  -cmd login -user alice -pass s3cr3t")
		fmt.Println("  -cmd add-note -user alice -title T -body B")
		fmt.Println("  -cmd list-notes -user alice")
	default:
		fmt.Println("unknown cmd; use -cmd help")
	}
}
