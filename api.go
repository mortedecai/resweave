package resweave

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
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

	keyPathHasSubSegment              = "pathHasSubSegment_%s"
	KeyURISegments                    = Key("URI_SEGMENTS")
	KeyRequestID                      = Key("INCOMING_REQUEST_ID")
	fmtResourceAlreadyExists          = "%w: '%s' in '%s'"
	fmtInstancedResourceAlreadyExists = "%w: '<id>/%s' in '%s'"
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
	ErrIDNotFound                 = errors.New("no ID found")
	ErrNilResource                = errors.New("cannot add a nil resource")
	ErrResourceAlreadyExists      = errors.New("sub-resource already exists")
	ErrChildResourceAlreadyExists = errors.New("child sub-resource already exists")
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
	// SetList sets the function to use for handling incoming list requests.
	SetList(f ResweaveFunc)
	// SetCreate sets the function to use for handling incoming create requests.
	SetCreate(f ResweaveFunc)
	// SetFetch sets the function to use for handling incoming fetch requests.
	SetFetch(f ResweaveFunc)
	// SetDelete sets the function to use for handling incoming delete requests.
	SetDelete(f ResweaveFunc)
	// SetUpdate sets the function to use for handling incoming update requests.
	SetUpdate(f ResweaveFunc)
	// SetID sets the regex for validating / parsing IDs for this resource.
	SetID(id ID) error
	// SetHandler sets the handler function for this resource.
	SetHandler(handler HandlerFunction)
	// GetIDValue retrieves the ID value from the provided context for this call.
	GetIDValue(ctx context.Context) (string, error)
	// AddResource adds a sub-resource which does not depend on any particular instantiated resource instance.
	//   For example, /users/search would allow the creation of searches across all users, but does not depend on any particular resource instantiations.
	//
	// Should the sub-resource need to access a particular resource instance, it MUST be added via AddChildResource.
	AddResource(Resource) error
	// AddChildResource adds a sub-resource which depends on a particular instantiated resource instance.
	//   For example:
	//     * `/users/<id>/profile` could allow the viewing or editing of a particular users profile.
	//     * `/users/<id>/emails` could allow the creation, viewing or editing of a particular users email(s).
	//     * `/users/<id>/emails/foo%40bar.com` could allow the deletion of a particular users `foo@bar.com` e-mail.
	AddChildResource(Resource) error
}

// BaseAPIRes supplies the basic building blocks for an APIResource.
// It may be used through composition
type BaseAPIRes struct {
	LogHolder
	name           ResourceName
	actionMap      actionFuncMap
	id             ID
	handler        HandlerFunction
	resources      ResourceMap
	childResources ResourceMap
}

// NewAPI creates a new APIResource instance with the provided name.
func NewAPI(name ResourceName) APIResource {
	bar := &BaseAPIRes{
		name:           name,
		LogHolder:      NewLogholder(name.String(), nil),
		actionMap:      make(actionFuncMap),
		resources:      make(ResourceMap),
		childResources: make(ResourceMap),
		id:             NumericID,
	}
	bar.SetHandler(bar.defaultHandler)
	return bar
}

func (bar *BaseAPIRes) Name() ResourceName {
	return bar.name
}

func (bar *BaseAPIRes) defaultFunction(_ context.Context, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (bar *BaseAPIRes) GetIDValue(ctx context.Context) (string, error) {
	return bar.GetResourceID(ctx, bar.name)
}

func (bar *BaseAPIRes) GetResourceID(ctx context.Context, name ResourceName) (string, error) {
	key := Key(fmt.Sprintf("id_%s", name.String()))
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

func (bar *BaseAPIRes) popSegmentPaths(ctx context.Context, idSegment int) (context.Context, []ResourceName) {
	uriSegments := ctx.Value(KeyURISegments).([]ResourceName)
	// housekeeping: If the last segment is empty, remove it.
	if len(uriSegments) > 0 && uriSegments[len(uriSegments)-1] == "" {
		uriSegments = uriSegments[:len(uriSegments)-1]
	}
	if idSegment < 0 {
		return ctx, uriSegments
	}
	if len(uriSegments) <= idSegment {
		return context.WithValue(ctx, KeyURISegments, []ResourceName{}), []ResourceName{}
	}
	uriSegments = uriSegments[(idSegment + 1):]
	return context.WithValue(ctx, KeyURISegments, uriSegments), uriSegments
}

func (bar *BaseAPIRes) storeID(c context.Context, req *http.Request) (context.Context, int, error) {
	const curMethod = "storeID"
	uriSegments, ok := c.Value(KeyURISegments).([]ResourceName)
	if !ok {
		return c, -1, errors.New("URI segments not found")
	}
	bar.Infow(curMethod, "Request URI", req.RequestURI, "Segment Count", len(uriSegments))

	ctx := c
	resourceNameIdx := -1
	idSegmentIdx := -1
	var idVal string = ""

	for i, v := range uriSegments {
		if resourceNameIdx >= 0 {
			idSegmentIdx = i
			idVal = string(v)
			break
		}
		if v == bar.Name() {
			resourceNameIdx = i
		}
	}

	bar.Infow(curMethod, "Segments", uriSegments, "idVal", idVal, "len(idVal)", len(idVal))

	if len(idVal) == 0 {
		return context.WithValue(ctx, Key(fmt.Sprintf(keyPathHasSubSegment, bar.name.String())), false), 0, nil
	}
	ctx = context.WithValue(ctx, Key(fmt.Sprintf(keyPathHasSubSegment, bar.name.String())), true)
	v, found := bar.id.Find(idVal)
	bar.Infow(curMethod, "idVal", idVal, "found", found, "v", v)
	if !found {
		// By allowing non-instanced sub-resources, it is now valid to have /resource/sub-resource as well as /resource/<id>/sub-resource.
		// This is a change from the original implementation, where only /resource/<id>/sub-resource was allowed.
		return ctx, resourceNameIdx, nil
	}

	ctx = context.WithValue(ctx, Key(fmt.Sprintf("id_%s", bar.name.String())), v)
	return ctx, idSegmentIdx, nil
}

func (bar *BaseAPIRes) hasID(ctx context.Context) bool {
	_, err := bar.GetIDValue(ctx)
	return err == nil
}

func (bar *BaseAPIRes) whichAction(ctx context.Context, httpMethod string) ActionType {
	switch httpMethod {
	case http.MethodGet:
		if bar.hasID(ctx) {
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

func (bar *BaseAPIRes) findSubResource(c context.Context, w http.ResponseWriter, req *http.Request) (Resource, error) {
	segments := c.Value(KeyURISegments).([]ResourceName)
	if len(segments) == 0 {
		return nil, errors.New("no segments found")
	}
	if !bar.hasID(c) {
		if res, found := bar.resources[segments[0]]; found {
			return res, nil
		}
		return nil, errors.New("no sub-resource found")
	}
	if res, found := bar.childResources[segments[0]]; found {
		return res, nil
	}
	return nil, errors.New("no instanced sub-resource found")
}

func (bar *BaseAPIRes) HandleCall(c context.Context, w http.ResponseWriter, req *http.Request) {
	ctx, idSegment, err := bar.storeID(c, req)
	if err != nil {
		bar.unknownResource(ctx, w, req)
		return
	}
	ctx, segments := bar.popSegmentPaths(ctx, idSegment)
	if len(segments) > 0 {
		// Not at the lowest level resource need to keep going.
		if res, err := bar.findSubResource(ctx, w, req); err != nil {
			bar.unknownResource(ctx, w, req)
		} else {
			res.HandleCall(ctx, w, req)
			return
		}
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

func (bar *BaseAPIRes) AddResource(r Resource) error {
	if r == nil {
		bar.Infow("AddResource", "Error", "resource was nil")
		return ErrNilResource
	}

	if _, found := bar.resources[r.Name()]; found {
		bar.Infow("AddResource", "Name", r.Name(), "Exists?", found)
		return fmt.Errorf(fmtResourceAlreadyExists, ErrResourceAlreadyExists, r.Name(), bar.Name())
	}
	bar.resources[r.Name()] = r
	bar.Infow("AddResource", "Name", fmt.Sprintf("'%s'", r.Name()), "Added", true)
	return nil
}

func (bar *BaseAPIRes) AddChildResource(r Resource) error {
	if r == nil {
		bar.Infow("AddChildResource", "Error", "resource was nil")
		return errors.New("cannot add a nil resource")
	}

	if _, found := bar.childResources[r.Name()]; found {
		bar.Infow("AddChildResource", "Name", r.Name(), "Exists?", found)
		return fmt.Errorf(fmtInstancedResourceAlreadyExists, ErrChildResourceAlreadyExists, r.Name(), bar.Name())
	}
	bar.childResources[r.Name()] = r
	bar.Infow("AddChildResource", "Name", fmt.Sprintf("'%s'", r.Name()), "Added", true)
	return nil
}
