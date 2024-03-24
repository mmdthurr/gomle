package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"mmd/gomle/db"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type FName struct {
	Name string
	Mime string
}

func ParseName(name string) FName {

	nf := strings.Split(name, ".")
	if len(nf) == 2 {
		return FName{
			Name: nf[0],
			Mime: nf[1],
		}

	} else {
		return FName{
			Name: nf[0],
			Mime: "",
		}
	}
}
func contains(s []string, e string) bool {
	e = strings.ToLower(e)
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func main() {
	up_key := flag.String("k", "upload", "upload path key")
	sqlitep := flag.String("db", "test.db", "sqlite database path ")
	tmpp := flag.String("tmp", "tmp", "tmp path ")
	tgToken := flag.String("tg", "123456", "telegram bot token")

	flag.Parse()

	Db := db.Connect(*sqlitep)

	AddForm := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>/mle/</title>
		<style>
			body {
				background-color: blanchedalmond;
				text-align: center;
			}
	
			.frm {
				max-width: 500px;
				margin: auto;
				text-align: center;
			}
	
			input {
				margin: 10px;
			}
	
			.mdia {
				display: block;
				max-width: 400px;
				margin-right: auto;
				margin-left: auto;
			}
		</style>
	</head>
	
	<body>
		<div class="frm">
			<form action="/mle/%s" method="post" enctype="multipart/form-data">
	
	
				<textarea dir="auto" name="meta" id="meta" cols="48" rows="4"></textarea>
				<input type="file" id="upfile" name="upfile">
				<input type="submit">
			</form>
		</div>
	
	</body>
	
	</html>
	`, *up_key)
	http.HandleFunc(fmt.Sprintf("/mle/%s", *up_key), func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {
			w.Write([]byte(AddForm))
			return
		}

		f, h, err := r.FormFile("upfile")
		if err != nil || h.Size > 20000000 {
			return
		}

		buffer := make([]byte, h.Size)
		_, err = f.Read(buffer)
		if err != nil {
			return
		}

		sha := sha256.New()
		sha.Write(buffer)
		shastr := hex.EncodeToString(sha.Sum(nil))

		pname := ParseName(h.Filename)
		newpost := db.Post{
			Sha256:   shastr,
			Meta:     r.FormValue("meta"),
			FileName: pname.Name,
			Mime:     pname.Mime,
		}
		Db.Create(&newpost)

		fm, err := os.Create(fmt.Sprintf("%s/%d", *tmpp, newpost.ID))
		if err != nil {
			return
		}
		defer fm.Close()
		fm.Write(buffer)

		tg_json, _ := json.Marshal(map[string]string{
			"chat_id":    "@naharlo",
			"parse_mode": "HTML",
			"text":       fmt.Sprintf(`<a href="http://ar3642.top/mle/post/%d">/mle/</a>`, newpost.ID),
		})

		http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", *tgToken), "application/json", bytes.NewBuffer(tg_json))

		w.Write([]byte("OK 200!"))

	})

	http.HandleFunc("/mle/post/", func(w http.ResponseWriter, r *http.Request) {
		up := r.URL.Path
		var post db.Post

		psp := strings.Split(up, "/")
		pid, err := strconv.Atoi(psp[len(psp)-1])
		if err != nil {
			w.Write([]byte("400 Bad Request"))
			return
		}

		if pid != 0 {
			err = Db.First(&post, pid).Error
			if err != nil {
				w.Write([]byte("404 Not Found"))
				return
			}
		}

		// w.Header().Set("Content-Type", "application/octet-stream")
		f, err := os.ReadFile(fmt.Sprintf("%s/%d", *tmpp, pid))
		if err != nil {
			w.Write([]byte("404 Not Found"))
			return
		}
		w.Write(f)

	})

	http.HandleFunc("/mle/", func(w http.ResponseWriter, r *http.Request) {
		up := r.URL.Path
		var posts []db.Post
		var page int

		if up == "/mle/" {
			page = 1
		} else {
			psp := strings.Split(up, "/")
			intp, err := strconv.Atoi(psp[len(psp)-1])
			if err != nil {
				w.Write([]byte("400 Bad Request"))
				return
			}
			page = intp
		}

		err := Db.Limit(10).Order("ID DESC").Offset(page*10 - 10).Find(&posts).Error
		if err != nil {
			w.Write([]byte("400 Bad Request"))
			return
		}

		var postshtml string

		for _, p := range posts {
			if contains([]string{"png", "jpeg", "webp", "jpg", "gif"}, p.Mime) {
				postshtml += fmt.Sprintf(`<a href="/mle/post/%d"> <img class="mdia" src="/mle/post/%d" alt="%s"></a>`, p.ID, p.ID, p.Meta)
			} else if contains([]string{"mp4", "webm"}, p.Mime) {
				postshtml += fmt.Sprintf(`<video class="mdia" src="/mle/post/%d" controls ></video>`, p.ID)
			} else if contains([]string{"mp3", "ogg"}, p.Mime) {
				postshtml += fmt.Sprintf(`<audio class="mdia" src="/mle/post/%d" controls ></audio>`, p.ID)
			} else {
				postshtml += fmt.Sprintf(`<a href="/mle/post/%d"> <img class="mdia" src="/mle/post/0" alt="%s"></a>`, p.ID, p.Meta)
			}

		}

		rsp := fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">

		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>/mle/</title>
			<style>
				body {
					background-color: blanchedalmond;
					text-align: center;
				}

				.frm {
					max-width: 500px;
					margin: auto;
					text-align: center;
				}

				input {
					margin: 10px;
				}

				.container {
					display: flex;
					flex-wrap: wrap;
					max-width: 1000px;
					margin: auto;
				}

				.mdia {
					display: block;
					max-width: 300px;
					margin-right: 5px;
					margin-left: 5px;
					margin-bottom: 10px;
				}
			</style>
		</head>

		<body>
			<div class="container">

			%s

			</div>
			<div style="text-align: center;">
				<a href="/mle/%d"> prev </a>
				<span> --- </span>
				<a href="/mle/%d"> next</a>
			</div>
		</body>

		</html>
		`, postshtml, page-1, page+1)
		w.Write([]byte(rsp))

	})

	http.ListenAndServe("127.0.0.1:5050", nil)
}
