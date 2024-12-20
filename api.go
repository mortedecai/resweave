package resweave

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// Key is a key type for looking up a resweave data in the context
type Key string

// ID is a regex representation of an ID format for an API resource.
type ID string

// actionType is a type alias for resweave actions
type ActionType int

// HandlerFunction is a type alias for the request handler
type HandlerFunction func(ActionType, context.Context, http.ResponseWriter, *http.Request)

// actionFuncMap is a type alias for a map of actionTypes to ResweaveFuncs
type actionFuncMap map[ActionType]ResweaveFunc

const (
	// NumericID is a default representation for a numeric identifier.
	NumericID ID = ID("([0-9]+)")
	UUIDv7    ID = ID(`^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$`)

	keyPathHasSubSegment = "pathHasSubSegment_%s"
	KeyRequestID         = Key("INCOMING_REQUEST_ID")
)

const (
	unknown ActionType = iota
	Create
	List
	Fetch
	Update
	Delete
)

var (
	ErrIDNotFound = errors.New("no ID found")
)

func (at ActionType) String() string {
	return [...]string{"Unknown", "Create", "List", "Fetch", "Update", "Delete"}[at]
}

// ID.IsValid returns true if the represented ID is valid regeix, or false and the error otherwise.
func (i ID) IsValid() (bool, error) {
	if _, err := regexp.Compile(string(i)); err != nil {
		return false, err
	}
	return true, nil
}

// ID.Find determines if an ID value matching the regex for this ID can be found in string s.
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
	LogHolder
	SetList(f ResweaveFunc)
	SetCreate(f ResweaveFunc)
	SetFetch(f ResweaveFunc)
	SetDelete(f ResweaveFunc)
	SetUpdate(f ResweaveFunc)
	SetID(id ID) error
	SetHandler(handler HandlerFunction)
	GetIDValue(ctx context.Context) (string, error)
}

// BaseAPIRes supplies the basic building blocks for an APIResource.
// It may be used through composition
type BaseAPIRes struct {
	LogHolder
	name      ResourceName
	actionMap actionFuncMap
	id        ID
	handler   HandlerFunction
}

// NewAPI creates a new APIResource instance with the provided name.
func NewAPI(name ResourceName) APIResource {
	bar := &BaseAPIRes{name: name, LogHolder: NewLogholder(name.String(), nil), actionMap: make(actionFuncMap)}
	bar.SetHandler(nil)
	return bar
}

func (bar *BaseAPIRes) Name() ResourceName {
	return bar.name
}

func (bar *BaseAPIRes) defaultFunction(_ context.Context, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (bar *BaseAPIRes) GetIDValue(ctx context.Context) (string, error) {
	key := Key(fmt.Sprintf("id_%s", bar.Name().String()))
	v := ctx.Value(key)
	var idStr string
	var ok bool
	if idStr, ok = v.(string); !ok {
		return "", ErrIDNotFound
	}

	return idStr, nil
}

func (bar *BaseAPIRes) SetID(id ID) error {
	if valid, err := id.IsValid(); !valid {
		return err
	}
	bar.id = id
	return nil
}

func (bar *BaseAPIRes) setFunction(at ActionType, f ResweaveFunc) {
	if f == nil {
		delete(bar.actionMap, at)
		return
	}
	bar.actionMap[at] = f
}

func (bar *BaseAPIRes) SetFetch(f ResweaveFunc) {
	bar.setFunction(Fetch, f)
}

func (bar *BaseAPIRes) SetDelete(f ResweaveFunc) {
	bar.setFunction(Delete, f)
}

func (bar *BaseAPIRes) SetList(f ResweaveFunc) {
	bar.setFunction(List, f)
}

func (bar *BaseAPIRes) SetCreate(f ResweaveFunc) {
	bar.setFunction(Create, f)
}

func (bar *BaseAPIRes) SetUpdate(f ResweaveFunc) {
	bar.setFunction(Update, f)
}

func (bar *BaseAPIRes) unknownResource(_ context.Context, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (bar *BaseAPIRes) storeID(c context.Context, req *http.Request) (context.Context, error) {
	const curMethod = "storeID"
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
		return context.WithValue(ctx, Key(fmt.Sprintf(keyPathHasSubSegment, bar.name.String())), false), nil
	}
	ctx = context.WithValue(ctx, Key(fmt.Sprintf(keyPathHasSubSegment, bar.name.String())), true)
	v, found := bar.id.Find(idVal)
	bar.Infow(curMethod, "idVal", idVal, "found", found, "v", v)
	if !found {
		// If the resource can't be found we have a problem; basically, a resource was specified but it's not valid
		// so we need to respond with an error.
		return ctx, ErrIDNotFound
	}

	ctx = context.WithValue(ctx, Key(fmt.Sprintf("id_%s", bar.name.String())), v)
	return ctx, nil
}

func (bar *BaseAPIRes) whichAction(ctx context.Context, httpMethod string) ActionType {
	switch httpMethod {
	case http.MethodGet:
		if _, err := bar.GetIDValue(ctx); err == nil {
			return Fetch
		}
		return List
	case http.MethodPost:
		return Create
	case http.MethodDelete:
		return Delete
	case http.MethodPut, http.MethodPatch:
		return Update
	default:
		return unknown
	}
}

func (bar *BaseAPIRes) SetHandler(handler HandlerFunction) {
	if handler == nil {
		bar.handler = bar.defaultHandler
		return
	}
	bar.handler = handler
}

func (bar *BaseAPIRes) HandleCall(c context.Context, w http.ResponseWriter, req *http.Request) {
	ctx, err := bar.storeID(c, req)
	if err != nil {
		bar.unknownResource(ctx, w, req)
		return
	}
	at := bar.whichAction(ctx, req.Method)
	if at == unknown {
		bar.defaultFunction(ctx, w, req)
		return
	}
	bar.handler(at, ctx, w, req)
}

func (bar *BaseAPIRes) defaultHandler(at ActionType, c context.Context, w http.ResponseWriter, req *http.Request) {
	var fun ResweaveFunc = bar.defaultFunction
	if f, found := bar.actionMap[at]; found {
		fun = f
	}
	if at == List {
		if hadChild := c.Value(Key(fmt.Sprintf(keyPathHasSubSegment, bar.name.String()))).(bool); hadChild {
			fun = bar.unknownResource
		}
	}
	fun(c, w, req)
}
