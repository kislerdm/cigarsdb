package cigarcentury

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Read(t *testing.T) {
	t.Skip("manual test")
	c := Client{HTTPClient: http.DefaultClient}
	r, err := c.Read(context.TODO(), "https://www.cigarcentury.com/en/cigars/arturo-fuente-casa-cuba-divine-inspiration")
	assert.NoError(t, err)
	assert.NotEmpty(t, r.AdditionalNotes)
}
