// Заготовка сервера. Закрывайте этапы по одному, пока не позеленеет go test ./tests/ -v
//
// Запуск: go run ./cmd/server — порт берётся из PORT, по умолчанию 8080.
// Панель уже раздаётся: откройте http://localhost:8080/ и смотрите, как этапы
// зеленеют по ходу работы.
package main

import (
	"log"
	"net/http"
	"os"

	"message-board/internal/handler"
	"message-board/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := store.NewStore()
	h := handler.NewHandlers(s)

	mux := http.NewServeMux()

	// Панель из frontend/. Каталог берётся относительно рабочего, поэтому
	// запускайте из корня модуля: go run ./cmd/server
	mux.Handle("/", http.FileServer(http.Dir("frontend")))

	// Этап 1: GET /health           -> 200, тело "ok"
	mux.HandleFunc("GET /health", h.HealthHandler)

	// Этап 2: POST /echo            -> тело запроса без изменений
	mux.HandleFunc("/echo", h.EchoHandler)

	// Этап 3: POST /echo            -> на application/json разобрать {"message": "..."} и вернуть JSON
	// (тот же хендлер /echo, но с проверкой Content-Type)

	// Этап 4: POST /messages        -> сохранить в памяти, 201
	// Этап 5: GET /messages         -> все сообщения, новые сверху
	mux.HandleFunc("/messages", h.MessagesHandler)

	// Этап 6: DELETE /messages/{id} -> 204, либо 404 если такого нет
	mux.HandleFunc("/messages/{id}", h.MessageByIDHandler)

	log.Printf("сервер слушает http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
