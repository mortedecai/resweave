package resweave

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type Key string
type ID string

const (
	NumericID ID = ID("([0-9]+)")
)

func (i ID) IsValid() (bool, error) {
	if _, err := regexp.Compile(string(i)); err != nil {
		return false, err
	}
	return true, nil
}

func (i ID) Find(s string) (string, bool) {
	var rxp *regexp.Regexp
	var err error

	if rxp, err = regexp.Compile(string(i)); err != nil {
		return "", false
	}

	if rxp.MatchString(s) {
		return rxp.FindString(s), true
	}

	return "", false
}

// APIResource is a basic APIResource which has a single point of entry for serving the supported access methods.
type APIResource interface {
	Resource
	logHolder
	SetList(f ResweaveFunc)
	SetCreate(f ResweaveFunc)
	SetID(id ID) error
	SetFetch(f ResweaveFunc)
}

// BaseAPIRes supplies the basic building blocks for an APIResource.
// It may be used through composition
type BaseAPIRes struct {
	logHolder
	name       ResourceName
	listFunc   ResweaveFunc
	createFunc ResweaveFunc
	fetchFunc  ResweaveFunc
	id         ID
}

// NewAPI creates a new APIResource instance with the provided name.
func NewAPI(name ResourceName) APIResource {
	bar := &BaseAPIRes{name: name, logHolder: newLogholder(name.String(), nil)}
	bar.listFunc = bar.defaultFunction
	bar.createFunc = bar.defaultFunction
	bar.fetchFunc = bar.defaultFunction
	return bar
}

func (bar *BaseAPIRes) Name() ResourceName {
	return bar.name
}

func (bar *BaseAPIRes) defaultFunction(_ context.Context, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (bar *BaseAPIRes) SetID(id ID) error {
	if valid, err := id.IsValid(); !valid {
		return err
	}
	bar.id = id
	return nil
}

func (bar *BaseAPIRes) SetFetch(f ResweaveFunc) {
	bar.fetchFunc = f
}

func (bar *BaseAPIRes) SetList(f ResweaveFunc) {
	bar.listFunc = f
}

func (bar *BaseAPIRes) SetCreate(f ResweaveFunc) {
	bar.createFunc = f
}

func (bar *BaseAPIRes) handleGet(c context.Context, w http.ResponseWriter, req *http.Request) {
	const curMethod = "handleGet"
	const methodKey = "method"
	uriSegments := strings.Split(req.URL.Path, "/")
	bar.Infow(curMethod, "Request URI", req.RequestURI, "Segment Count", len(uriSegments))

	ctx := c
	uriSeg := -1
	var idVal string = ""

	for i, v := range uriSegments {
		if uriSeg >= 0 {
			idVal = v
			break
		}
		if v == bar.Name().String() {
			uriSeg = i
		}
	}

	bar.Infow(curMethod, "Segments", uriSegments, "idVal", idVal, "len(idVal)", len(idVal))

	if len(idVal) == 0 {
		bar.Infow(curMethod, methodKey, "LIST")
		bar.listFunc(ctx, w, req)
		return
	}
	v, found := bar.id.Find(idVal)
	bar.Infow(curMethod, "idVal", idVal, "found", found, "v", v)
	if found {
		ctx := context.WithValue(ctx, Key(fmt.Sprintf("id_%s", bar.name.String())), v)
		bar.Infow(curMethod, methodKey, "FETCH")
		bar.fetchFunc(ctx, w, req)
		return
	} else {
		// TODO:  Will need to check here for nested API resource when added to code.
		bar.Infow(curMethod, methodKey, "NONE")
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (bar *BaseAPIRes) HandleCall(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		bar.handleGet(ctx, w, req)
	case http.MethodPost:
		bar.createFunc(ctx, w, req)
	default:
		bar.defaultFunction(ctx, w, req)
	}
}
