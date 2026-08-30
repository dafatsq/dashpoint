package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type strictParseTarget struct {
	Name string `json:"name"`
}

func runStrictParse(t *testing.T, body string) (*strictParseTarget, error) {
	t.Helper()
	app := fiber.New()
	var captured *strictParseTarget
	app.Post("/", func(c *fiber.Ctx) error {
		var req strictParseTarget
		if err := parseStrictJSONBody(c, &req, productMaxJSONBodyBytes); err != nil {
			return err
		}
		captured = &req
		return c.SendStatus(fiber.StatusOK)
	})
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		return nil, errStrictTestResponse{status: resp.StatusCode}
	}
	return captured, nil
}

type errStrictTestResponse struct{ status int }

func (e errStrictTestResponse) Error() string { return string(rune(e.status)) }

func TestParseStrictJSONBodyAcceptsValidSingleObject(t *testing.T) {
	req, err := runStrictParse(t, `{"name":"widget"}`)
	if err != nil {
		t.Fatalf("expected valid body to parse, got %v", err)
	}
	if req.Name != "widget" {
		t.Fatalf("expected decoded name, got %q", req.Name)
	}
}

func TestParseStrictJSONBodyRejectsUnknownFields(t *testing.T) {
	if _, err := runStrictParse(t, `{"name":"widget","admin":true}`); err != nil {
		if _, isStatus := err.(errStrictTestResponse); !isStatus {
			t.Fatalf("expected rejection at response level, got %v", err)
		}
		return
	}
	t.Fatalf("expected unknown field to be rejected")
}

func TestParseStrictJSONBodyRejectsOversizeAndGarbage(t *testing.T) {
	if _, err := runStrictParse(t, `{"name":"`+strings.Repeat("x", productMaxJSONBodyBytes)+`"}`); err == nil {
		t.Fatalf("expected oversize body to be rejected")
	}
	if _, err := runStrictParse(t, `{"name":"widget"} trailing`); err == nil {
		t.Fatalf("expected trailing content to be rejected")
	}
	if _, err := runStrictParse(t, ``); err == nil {
		t.Fatalf("expected empty body to be rejected")
	}
}
