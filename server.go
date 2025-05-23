package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/kkdai/youtube/v2"
)

func get_videos_from_playlist(list_url string) ([]*youtube.PlaylistEntry, error) {
	client := youtube.Client{}
	// Get playlist details
	playlist, err := client.GetPlaylist(list_url)
	if err != nil {
		return nil, err
	}
	return playlist.Videos, nil
}

func main() {
	var Videos []*youtube.PlaylistEntry

	file_server := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
		} else {
			file_server.ServeHTTP(w, r)
		}
	})

	http.HandleFunc("/add_list", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			url := r.FormValue("playlist")
			vids, err := get_videos_from_playlist(url)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			Videos = vids
			io.WriteString(w, "COMPLETED")
		}
	})

	http.HandleFunc("/get_list", func(w http.ResponseWriter, r *http.Request) {
		type vids struct {
			Id    string
			Title string
		}
		vides := make([]vids, 0)
		for _, v := range Videos {
			vides = append(vides, vids{Id: v.ID, Title: v.Title})
		}
		// w.Header().Add("Content-Type", "Application/json")
		// json.NewEncoder(w).Encode(vides)
		temp, err := template.ParseFiles("watch.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = temp.Execute(w, vides)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/thumbnail/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		resp, err := http.Get(fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", id))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "image/jpeg")
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	http.ListenAndServe(":8080", nil)
}
