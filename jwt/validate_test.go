package jwt_test

import (
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/internal/json"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGHIssue10(t *testing.T) {
	t.Parallel()
	t.Run(jwt.IssuerKey, func(t *testing.T) {
		t.Parallel()
		t1 := jwt.New()
		t1.Set(jwt.IssuerKey, "github.com/lestrrat-go/jwx")

		// This should succeed, because WithIssuer is not provided in the
		// optional parameters
		if !assert.NoError(t, jwt.Validate(t1), "t1.Validate should succeed") {
			return
		}

		// This should succeed, because WithIssuer is provided with same value
		if !assert.NoError(t, jwt.Validate(t1, jwt.WithIssuer(t1.Issuer())), "t1.Validate should succeed") {
			return
		}

		if !assert.Error(t, jwt.Validate(t1, jwt.WithIssuer("poop")), "t1.Validate should fail") {
			return
		}
	})
	t.Run(jwt.AudienceKey, func(t *testing.T) {
		t.Parallel()
		t1 := jwt.New()
		err := t1.Set(jwt.AudienceKey, []string{"foo", "bar", "baz"})
		if err != nil {
			t.Fatalf("Failed to set audience claim: %s", err.Error())
		}

		// This should succeed, because WithAudience is not provided in the
		// optional parameters
		err = jwt.Validate(t1)
		if err != nil {
			t.Fatalf("Error verifying claim: %s", err.Error())
		}

		// This should succeed, because WithAudience is provided, and its
		// value matches one of the audience values
		if !assert.NoError(t, jwt.Validate(t1, jwt.WithAudience("baz")), "token.Validate should succeed") {
			return
		}

		if !assert.Error(t, jwt.Validate(t1, jwt.WithAudience("poop")), "token.Validate should fail") {
			return
		}
	})
	t.Run(jwt.SubjectKey, func(t *testing.T) {
		t.Parallel()
		t1 := jwt.New()
		t1.Set(jwt.SubjectKey, "github.com/lestrrat-go/jwx")

		// This should succeed, because WithSubject is not provided in the
		// optional parameters
		if !assert.NoError(t, jwt.Validate(t1), "token.Validate should succeed") {
			return
		}

		// This should succeed, because WithSubject is provided with same value
		if !assert.NoError(t, jwt.Validate(t1, jwt.WithSubject(t1.Subject())), "token.Validate should succeed") {
			return
		}

		if !assert.Error(t, jwt.Validate(t1, jwt.WithSubject("poop")), "token.Validate should fail") {
			return
		}
	})
	t.Run(jwt.NotBeforeKey, func(t *testing.T) {
		t.Parallel()
		t1 := jwt.New()

		// NotBefore is set to future date
		tm := time.Now().Add(72 * time.Hour)
		t1.Set(jwt.NotBeforeKey, tm)

		// This should fail, because nbf is the future
		if !assert.Error(t, jwt.Validate(t1), "token.Validate should fail") {
			return
		}

		// This should succeed, because we have given reaaaaaaly big skew
		// that is well enough to get us accepted
		if !assert.NoError(t, jwt.Validate(t1, jwt.WithAcceptableSkew(73*time.Hour)), "token.Validate should succeed") {
			return
		}

		// This should succeed, because we have given a time
		// that is well enough into the future
		if !assert.NoError(t, jwt.Validate(t1, jwt.WithClock(jwt.ClockFunc(func() time.Time { return tm.Add(time.Hour) }))), "token.Validate should succeed") {
			return
		}
	})
	t.Run(jwt.ExpirationKey, func(t *testing.T) {
		t.Parallel()
		t1 := jwt.New()

		// issuedat = 1 Hr before current time
		tm := time.Now()
		t1.Set(jwt.IssuedAtKey, tm.Add(-1*time.Hour))

		// valid for 2 minutes only from IssuedAt
		t1.Set(jwt.ExpirationKey, tm.Add(-58*time.Minute))

		// This should fail, because exp is set in the past
		if !assert.Error(t, jwt.Validate(t1), "token.Validate should fail") {
			return
		}

		// This should succeed, because we have given big skew
		// that is well enough to get us accepted
		if !assert.NoError(t, jwt.Validate(t1, jwt.WithAcceptableSkew(time.Hour)), "token.Validate should succeed (1)") {
			return
		}

		// This should succeed, because we have given a time
		// that is well enough into the past
		clock := jwt.ClockFunc(func() time.Time {
			return tm.Add(-59 * time.Minute)
		})
		if !assert.NoError(t, jwt.Validate(t1, jwt.WithClock(clock)), "token.Validate should succeed (2)") {
			return
		}
	})
	t.Run("Parse and validate", func(t *testing.T) {
		t.Parallel()
		t1 := jwt.New()

		// issuedat = 1 Hr before current time
		tm := time.Now()
		t1.Set(jwt.IssuedAtKey, tm.Add(-1*time.Hour))

		// valid for 2 minutes only from IssuedAt
		t1.Set(jwt.ExpirationKey, tm.Add(-58*time.Minute))

		buf, err := json.Marshal(t1)
		if !assert.NoError(t, err, `json.Marshal should succeed`) {
			return
		}

		_, err = jwt.Parse(buf, jwt.WithValidate(true))
		// This should fail, because exp is set in the past
		if !assert.Error(t, err, "jwt.Parse should fail") {
			return
		}

		_, err = jwt.Parse(buf, jwt.WithValidate(true), jwt.WithAcceptableSkew(time.Hour))
		// This should succeed, because we have given big skew
		// that is well enough to get us accepted
		if !assert.NoError(t, err, "jwt.Parse should succeed (1)") {
			return
		}

		// This should succeed, because we have given a time
		// that is well enough into the past
		clock := jwt.ClockFunc(func() time.Time {
			return tm.Add(-59 * time.Minute)
		})
		_, err = jwt.Parse(buf, jwt.WithValidate(true), jwt.WithClock(clock))
		if !assert.NoError(t, err, "jwt.Parse should succeed (2)") {
			return
		}
	})
	t.Run("any claim value", func(t *testing.T) {
		t.Parallel()
		t1 := jwt.New()
		t1.Set("email", "email@example.com")

		// This should succeed, because WithClaimValue("email", "xxx") is not provided in the
		// optional parameters
		if !assert.NoError(t, jwt.Validate(t1), "t1.Validate should succeed") {
			return
		}

		// This should succeed, because WithClaimValue is provided with same value
		if !assert.NoError(t, jwt.Validate(t1, jwt.WithClaimValue("email", "email@example.com")), "t1.Validate should succeed") {
			return
		}

		if !assert.Error(t, jwt.Validate(t1, jwt.WithClaimValue("email", "poop")), "t1.Validate should fail") {
			return
		}
		if !assert.Error(t, jwt.Validate(t1, jwt.WithClaimValue("xxxx", "email@example.com")), "t1.Validate should fail") {
			return
		}
		if !assert.Error(t, jwt.Validate(t1, jwt.WithClaimValue("xxxx", "")), "t1.Validate should fail") {
			return
		}
	})
}
