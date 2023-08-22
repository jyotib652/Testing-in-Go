package main

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	var theTests = []struct {
		name                    string
		url                     string
		expectedStatusCode      int
		expectedURL             string
		expectedFirstStatusCode int
	}{
		{"home", "/", http.StatusOK, "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound, "/fish", http.StatusNotFound},
		{"profile", "/user/profile", http.StatusOK, "/", http.StatusTemporaryRedirect},
	}

	routes := app.routes()

	// create a test server(web server)
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	// pathToTemplates = "./../../templates/"

	// range through the test data
	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		// in ts.Client(), we're calling test server with built-in http client which is a part of that test server.
		// but we need to create our own client. Because by default http client will follow upto 10 redirects.
		// That's the maximum it will follow and it will always return the last one. We can get around that by creating
		// second client to be used only for this particular test case.
		// So let's create the second client. This will be a custom client:

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // this way it'll accept invalid SSL certificates (which aren't signed)
		}

		client := &http.Client{
			Transport: tr,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // only this way we'll get the first response code
			},
		}

		// For redirect or temporary redirect you get temporary redirect status code
		// and final status code.
		// here we're testing for final status code.
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s: expected status %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}

		if resp.Request.URL.Path != e.expectedURL {
			t.Errorf("%s: expected final url of %s but got: %s", e.name, e.expectedURL, resp.Request.URL.Path)
		}

		resp2, _ := client.Get(ts.URL + e.url)
		if resp2.StatusCode != e.expectedFirstStatusCode {
			t.Errorf("%s expected first returned status code to be %d but got: %d", e.name, e.expectedFirstStatusCode, resp2.StatusCode)
		}
	}
}

// func Test_App_HomeOld(t *testing.T) {
// 	// here, we want to call the handler directly, nit by creating a server and then call the handler
// 	// we want to test obly the handler: so first,
// 	// crate a request
// 	req, _ := http.NewRequest("GET", "/", nil)

// 	// Now add the value to the conext and add the session information to the context
// 	req = addContextAndSessionToRequest(req, app)

// 	// we also need a response, so we will create variable 'rr' response recorder which will serve as response writer
// 	// response recorder is not response writer but it satifies the interface for response writer
// 	rr := httptest.NewRecorder()

// 	// now, we create the handler
// 	handler := http.HandlerFunc(app.Home)

// 	// now, we serve the request
// 	handler.ServeHTTP(rr, req)

// 	// check status code
// 	if rr.Code != http.StatusOK {
// 		t.Errorf("Test_App_Home returned wrong status code; expected 200 but got %d", rr.Code)
// 	}

// 	body, _ := io.ReadAll(rr.Body)
// 	if !strings.Contains(string(body), `<small>From Session:`) {
// 		t.Error("did not find correct text from html")
// 	}
// }

func Test_app_Home(t *testing.T) {
	var tests = []struct {
		name         string
		putInSession string
		expectedHTML string
	}{
		{"first visit", "", "<small>From Session:"},
		{"second visit", "hello, world!", "<small>From Session: hello, world!"},
	}

	for _, e := range tests {
		// here, we want to call the handler directly, nit by creating a server and then call the handler
		// we want to test obly the handler: so first,
		// crate a request
		req, _ := http.NewRequest("GET", "/", nil)

		// Now add the value to the conext and add the session information to the context
		req = addContextAndSessionToRequest(req, app)

		// make sure the session is empty
		_ = app.Session.Destroy(req.Context())

		if e.putInSession != "" {
			app.Session.Put(req.Context(), "test", e.putInSession)
		}

		// we also need a response, so we will create variable 'rr' response recorder which will serve as response writer
		// response recorder is not response writer but it satifies the interface for response writer
		rr := httptest.NewRecorder()

		// now, we create the handler
		handler := http.HandlerFunc(app.Home)

		// now, we serve the request
		handler.ServeHTTP(rr, req)

		// check status code
		if rr.Code != http.StatusOK {
			t.Errorf("Test_App_Home returned wrong status code; expected 200 but got %d", rr.Code)
		}

		body, _ := io.ReadAll(rr.Body)
		if !strings.Contains(string(body), e.expectedHTML) {
			t.Errorf("%s did not find %s in response body", e.name, e.expectedHTML)
		}
	}
}

func Test_app_renderWithBadTemplate(t *testing.T) {
	// set templatepath to a location with a bad template
	pathToTemplates = "./testdata/"

	req, _ := http.NewRequest("GET", "/", nil)
	req = addContextAndSessionToRequest(req, app)
	rr := httptest.NewRecorder()

	err := app.render(rr, req, "bad.page.gohtml", &TemplateData{})
	if err == nil {
		t.Error("expected an error from bad template(bad.page.gohtml) but did not get one")
	}

	// set the templates back to the original location
	pathToTemplates = "./../../templates/"
}

func getCtx(r *http.Request) context.Context {
	ctx := context.WithValue(r.Context(), contextUserKey, "unknown")
	return ctx
}

func addContextAndSessionToRequest(req *http.Request, app application) *http.Request {
	// build a request with context
	req = req.WithContext(getCtx(req))

	// now, add information for the session
	// "X-Session" is expected to be in the request if we are going to be able to use session
	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session"))

	return req.WithContext(ctx)
}

func Test_app_Login(t *testing.T) {
	var tests = []struct {
		name               string
		postedData         url.Values
		expectedStatusCode int
		expectedLoc        string // for login page or profile page url
	}{
		{
			name: "valid login",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/user/profile",
		},
		{
			name: "missing form data",
			postedData: url.Values{
				"email":    {""},
				"password": {""},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
		{
			name: "user not found",
			postedData: url.Values{
				"email":    {"you@there.com"},
				"password": {"password"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
		{
			name: "bad credentials",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"password"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
	}

	for _, e := range tests {
		// bild a request
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(e.postedData.Encode()))
		req = addContextAndSessionToRequest(req, app)
		// We have to set the header to the same thing that a form post does in a html page
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // that's the content type that Go expects to find from an HTML form post
		rr := httptest.NewRecorder()                                        // response recorder
		handler := http.HandlerFunc(app.Login)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: returned wrong status code; expected %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		actualLoc, err := rr.Result().Location()
		// t.Log("actual location:", actualLoc.String())
		if err == nil {
			if actualLoc.String() != e.expectedLoc {
				t.Errorf("%s: expected location %s but got %s", e.name, e.expectedLoc, actualLoc.String())
			}
		} else {
			t.Errorf("%s: no location header is set", e.name)
		}
	}
}
