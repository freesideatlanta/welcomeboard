package photowelcome

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"os"
	"strings"
	"text/template"
)

//go:embed greetings
var greetings embed.FS

//go:embed templates/photo.template.html
var photoPageTemplateRaw string

var photoTemplateParsed *template.Template

type ImageInfo struct {
	Mime string
	Data string
}

func init() {
	photoTemplateParsed = template.Must(template.New("photo_page").Parse(photoPageTemplateRaw))
}

func PhotoGreeting() []byte {
	items, err := greetings.ReadDir("greetings")
	if err != nil {
		fmt.Println("oops4: ", err.Error())
	}
	n := rand.Int() % len(items)
	image := items[n]

	splitName := strings.Split(image.Name(), ".")
	extension := splitName[len(splitName)-1]
	mimeType := mime.TypeByExtension(extension)
	f, err := greetings.Open("greetings" + string(os.PathSeparator) + image.Name())
	if err != nil {
		fmt.Println("oops5: ", err.Error())
	}
	defer f.Close()
	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, f)
	if err != nil {
		fmt.Println("oops6: ", err.Error())
	}
	imgenc := base64.StdEncoding.EncodeToString(buf.Bytes())
	buf2 := bytes.NewBuffer([]byte{})
	photoTemplateParsed.Execute(buf2, ImageInfo{
		Mime: mimeType,
		Data: imgenc,
	})
	return buf2.Bytes()

}
