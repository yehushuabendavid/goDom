package test

import (
	"goDom"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

type jar struct {
	lock sync.Mutex
	data map[string][]*http.Cookie
}

func (j *jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.lock.Lock()
	j.data[u.Host] = cookies
	j.lock.Unlock()
}
func (j *jar) Cookies(u *url.URL) []*http.Cookie {
	j.lock.Lock()
	cookies := j.data[u.Host]
	j.lock.Unlock()
	return cookies
}
func newJar() *jar {
	j := &jar{data: map[string][]*http.Cookie{}}
	return j
}

func main() {
	n := http.NewServeMux()
	n.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<b>Salut<!-- ok
		cool <b> toto <b> 
		<br>
		-->  </b>`))
	})
	n.HandleFunc("/toto", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><body>
		<b>
<br>
</b><!-- Comment <br> -->
		<script>$$$$$$$$$$$$$$$
   scr "</script>" 1>2
if 1<1 coule
asd
"toto\"titi\" "
</script></html>
		`))
	})
	n.HandleFunc("/classic", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><head><title>Classic</title></head><body><h1 id="main" class="error" selected zoom=1 onclick="action(\"toto\")">Hi <b><b>man</b></b></h1><i></i><test></test></body></html>`))
	})
	go http.ListenAndServe("0.0.0.0:8090", n)

	for i := 0; i < 100; i++ {
		fmt.Println(".")
	}
	fmt.Println("Start")
	client := http.Client{}
	client.Jar = newJar()
	r, err := client.Get(
		"https://devdinocdn.com/mako/election_graphs/Home/Official?white=1")
	if err == nil {
		b, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		doc := dom.DocumentParser(string(b))
		for _,e:= range doc.Selector(".graph-unit"){
			fmt.Println(
				e.Selector(".graph-result-index")[0].InnerContent(),
				html.UnescapeString( e.Selector("img")[0].Attr["alt"]))
		}

	} else {
		fmt.Println(err)
	}
}
