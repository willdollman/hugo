// Copyright 2016 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
	"bytes"
	"html"

	"github.com/miekg/mmark"
	"github.com/willdollman/blackfriday"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

type LinkResolverFunc func(ref string) (string, error)
type FileResolverFunc func(ref string) (string, error)

// HugoHTMLRenderer wraps a blackfriday.Renderer, typically a blackfriday.Html
// Enabling Hugo to customise the rendering experience
type HugoHTMLRenderer struct {
	*RenderingContext
	blackfriday.Renderer
}

func (renderer *HugoHTMLRenderer) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	if viper.GetBool("pygmentsCodeFences") && (lang != "" || viper.GetBool("pygmentsCodeFencesGuessSyntax")) {
		opts := viper.GetString("pygmentsOptions")
		str := html.UnescapeString(string(text))
		out.WriteString(Highlight(str, lang, opts))
	} else {
		renderer.Renderer.BlockCode(out, text, lang)
	}
}

func (renderer *HugoHTMLRenderer) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	if renderer.LinkResolver == nil || bytes.HasPrefix(link, []byte("HAHAHUGOSHORTCODE")) {
		// Use the blackfriday built in Link handler
		renderer.Renderer.Link(out, link, title, content)
	} else {
		// set by SourceRelativeLinksEval
		newLink, err := renderer.LinkResolver(string(link))
		if err != nil {
			newLink = string(link)
			jww.ERROR.Printf("LinkResolver: %s", err)
		}
		renderer.Renderer.Link(out, []byte(newLink), title, content)
	}
}
func (renderer *HugoHTMLRenderer) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	if renderer.FileResolver == nil || bytes.HasPrefix(link, []byte("HAHAHUGOSHORTCODE")) {
		// Use the blackfriday built in Image handler
		renderer.Renderer.Image(out, link, title, alt)
	} else {
		// set by SourceRelativeLinksEval
		newLink, err := renderer.FileResolver(string(link))
		if err != nil {
			newLink = string(link)
			jww.ERROR.Printf("FileResolver: %s", err)
		}
		renderer.Renderer.Image(out, []byte(newLink), title, alt)
	}
}

// ListItem adds task list support to the Blackfriday renderer.
func (renderer *HugoHTMLRenderer) ListItem(out *bytes.Buffer, text []byte, flags int) {
	if !renderer.Config.TaskLists {
		renderer.Renderer.ListItem(out, text, flags)
		return
	}

	switch {
	case bytes.HasPrefix(text, []byte("[ ] ")):
		text = append([]byte(`<input type="checkbox" disabled class="task-list-item">`), text[3:]...)

	case bytes.HasPrefix(text, []byte("[x] ")) || bytes.HasPrefix(text, []byte("[X] ")):
		text = append([]byte(`<input type="checkbox" checked disabled class="task-list-item">`), text[3:]...)
	}

	renderer.Renderer.ListItem(out, text, flags)
}

// List adds task list support to the Blackfriday renderer.
func (renderer *HugoHTMLRenderer) List(out *bytes.Buffer, text func() bool, flags int) {
	if !renderer.Config.TaskLists {
		renderer.Renderer.List(out, text, flags)
		return
	}
	marker := out.Len()
	renderer.Renderer.List(out, text, flags)
	if out.Len() > marker {
		list := out.Bytes()[marker:]
		if bytes.Contains(list, []byte("task-list-item")) {
			// Rewrite the buffer from the marker
			out.Truncate(marker)
			// May be either dl, ul or ol
			list := append(list[:4], append([]byte(` class="task-list"`), list[4:]...)...)
			out.Write(list)
		}
	}
}

// HugoMmarkHTMLRenderer wraps a mmark.Renderer, typically a mmark.html
// Enabling Hugo to customise the rendering experience
type HugoMmarkHTMLRenderer struct {
	mmark.Renderer
}

func (renderer *HugoMmarkHTMLRenderer) BlockCode(out *bytes.Buffer, text []byte, lang string, caption []byte, subfigure bool, callouts bool) {
	if viper.GetBool("pygmentsCodeFences") && (lang != "" || viper.GetBool("pygmentsCodeFencesGuessSyntax")) {
		str := html.UnescapeString(string(text))
		out.WriteString(Highlight(str, lang, ""))
	} else {
		renderer.Renderer.BlockCode(out, text, lang, caption, subfigure, callouts)
	}
}
