package util

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mmmorris1975/simple-logger"
	"os"
	"reflect"
	"testing"
)

type MockRoleGetter struct {
	r Roles
}

func NewMockRoleGetter(r []string) RoleGetter {
	return &MockRoleGetter{r: Roles(r)}
}

func (m *MockRoleGetter) Roles() Roles {
	return m.r.Dedup()
}

func ExampleRoleGetter() {
	roles := []string{
		"mock3", "mock2", "mock1", "mock2", "mock4", "mock1",
	}
	m := NewMockRoleGetter(roles)
	for _, r := range m.Roles() {
		fmt.Println(r)
	}
	// Output:
	// mock1
	// mock2
	// mock3
	// mock4
}

func TestEmptyRoleGetter(t *testing.T) {
	m := NewMockRoleGetter([]string{})
	r := m.Roles()

	t.Logf("Empty role result: %v", r)
	if len(r) != 0 {
		t.Errorf("Found unexpected roles from empty input")
	}
}

func TestNilRoleGetter(t *testing.T) {
	m := NewMockRoleGetter(nil)
	r := m.Roles()

	t.Logf("Nil role result: %v", r)
	if len(r) != 0 {
		t.Errorf("Found unexpected roles from nil input")
	}
}

func TestNewAwsRoleGetterDefault(t *testing.T) {
	defer func() {
		if x := recover(); x != nil {
			t.Errorf("panic() from default NewAwsRoleGetter()")
		}
	}()
	s, err := session.NewSessionWithOptions(session.Options{})
	if err != nil {
		t.Errorf("Error creating AWS session: %v", err)
	}
	NewAwsRoleGetter(s, "u")
}

func TestParsePolicy(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		j := `{"Statement": [{"Effect": "None"}, {"Effect": "Deny", "Action": "sts:AssumeRole"}, 
                             {"Effect": "Allow", "Action": ["sts:AssumeRole", "s3:*"], "Resource": "a"},
                             {"Effect": "Allow", "Action": "sts:AssumeRole", "Resource": "x"},
                             {"Effect": "Allow", "Action": "sts:AssumeRole", "Resource": ["y", "z"]}
                            ]}`
		r, err := parsePolicy(&j)
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(r, Roles([]string{"a", "x", "y", "z"})) {
			t.Errorf("unexpected Roles value")
		}
	})

	t.Run("string", func(t *testing.T) {
		r, err := parsePolicy(aws.String("string-only"))
		if err != nil {
			t.Error(err)
			return
		}

		if len(r) > 0 {
			t.Error("roles size was > 0")
		}
	})

	t.Run("string array", func(t *testing.T) {
		r, err := parsePolicy(aws.String(`{"Statement": ["a", "b", "c"]}`))
		if err != nil {
			t.Error(err)
			return
		}

		if len(r) > 0 {
			t.Error("roles size was > 0")
		}
	})

	t.Run("bad map", func(t *testing.T) {
		r, err := parsePolicy(aws.String(`{"Statement": [{}, {"b": 1}]}`))
		if err != nil {
			t.Error(err)
			return
		}

		if len(r) > 0 {
			t.Error("roles size was > 0")
		}
	})

	t.Run("nil doc", func(t *testing.T) {
		if _, err := parsePolicy(nil); err == nil {
			t.Error("did not see expected error")
			return
		}
	})

	t.Run("empty doc", func(t *testing.T) {
		r, err := parsePolicy(aws.String(""))
		if err != nil {
			t.Error(err)
			return
		}

		if len(r) > 0 {
			t.Error("roles size was > 0")
		}
	})
}

func Example_debugNilClient() {
	r := NewAwsRoleGetter(nil, "u")
	r.debug("test")
	// Output:
	//
}

func Example_debugAwsLogger() {
	l := aws.LoggerFunc(func(v ...interface{}) { fmt.Fprintln(os.Stdout, v...) })
	c := new(aws.Config).WithLogger(l).WithLogLevel(aws.LogDebug)
	s := session.Must(session.NewSession(c))
	r := NewAwsRoleGetter(s, "u").WithLogger(l)
	r.debug("test")
	// Output:
	// test
}

func Example_debugSimpleLogger() {
	l := simple_logger.NewLogger(os.Stdout, "", 0)
	l.SetLevel(simple_logger.DEBUG)
	c := new(aws.Config).WithLogger(l).WithLogLevel(aws.LogDebug)
	s := session.Must(session.NewSession(c))
	r := NewAwsRoleGetter(s, "u").WithLogger(l)
	r.debug("test")
	// Output:
	// DEBUG test
}

func Example_errorNilLogger() {
	r := NewAwsRoleGetter(nil, "u")
	r.error("test")
	// Output:
	//
}

func Example_errorAwsLogger() {
	l := aws.LoggerFunc(func(v ...interface{}) { fmt.Fprintln(os.Stdout, v...) })
	c := new(aws.Config).WithLogger(l).WithLogLevel(aws.LogDebug)
	s := session.Must(session.NewSession(c))
	r := NewAwsRoleGetter(s, "u").WithLogger(l)
	r.error("test")
	// Output:
	// test
}

func Example_errorSimpleLogger() {
	l := simple_logger.NewLogger(os.Stdout, "", 0)
	l.SetLevel(simple_logger.INFO)
	c := new(aws.Config).WithLogger(l).WithLogLevel(aws.LogDebug)
	s := session.Must(session.NewSession(c))
	r := NewAwsRoleGetter(s, "u").WithLogger(l)
	r.error("test")
	// Output:
	// ERROR test
}
