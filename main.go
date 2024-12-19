package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Task struct {
	ID          uint      `gorm:"primaryKey" json:"ID"`
	UserID      uint      `json:"UserID"`
	Description string    `json:"Description"`
	RecallDate  time.Time `json:"RecallDate"`
	CreatedDate time.Time `json:"CreatedDate"`
	User        User      `gorm:"foreignKey:UserID" json:"User"`
}

var db *gorm.DB

func initDB() {
	dsn := "user=task_user password=password123 dbname=task_manager host=localhost port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}

	err = db.AutoMigrate(&User{}, &Task{})
	if err != nil {
		log.Fatal("Ошибка миграции:", err)
	}

	fmt.Println("Успешное подключение к базе данных!")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	db.Create(&user)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	db.Find(&users)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
	}
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	db.Model(&User{}).Where("id = ?", user.ID).Updates(user)
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated successfully"})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	db.Delete(&User{}, user.ID)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Println("Ошибка при декодировании JSON:", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("Создание задачи: %+v\n", task)
	task.CreatedDate = time.Now()
	db.Create(&task)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task created successfully"})
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []Task
	db.Preload("User").Find(&tasks) // Preload для подгрузки пользователя
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(tasks)
	if err != nil {
		log.Println("Ошибка кодирования JSON:", err)
	}
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Обновляем только поля задачи, которые были изменены
	result := db.Model(&Task{}).Where("id = ?", task.ID).Updates(Task{
		UserID:      task.UserID,
		Description: task.Description,
		RecallDate:  task.RecallDate,
	})

	if result.RowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Task updated successfully"})
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	db.Delete(&Task{}, task.ID)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted successfully"})
}

func renderHTML(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func main() {
	initDB()

	http.HandleFunc("/", renderHTML)
	http.HandleFunc("/users/create", createUser)
	http.HandleFunc("/users", getUsers)
	http.HandleFunc("/users/update", updateUser)
	http.HandleFunc("/users/delete", deleteUser)

	http.HandleFunc("/tasks", getTasks)
	http.HandleFunc("/tasks/create", createTask)
	http.HandleFunc("/tasks/update", updateTask)
	http.HandleFunc("/tasks/delete", deleteTask)

	port := ":8080"
	fmt.Println("Сервер запущен на порту", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
