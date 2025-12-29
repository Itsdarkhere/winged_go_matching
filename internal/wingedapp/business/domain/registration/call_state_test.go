package registration_test

import (
	"context"
	"fmt"
	"testing"
	"wingedapp/pgtester/internal/util/strutil"
	"wingedapp/pgtester/internal/wingedapp/business/domain/registration"
	"wingedapp/pgtester/internal/wingedapp/testsuite"

	"github.com/aarondl/null/v8"
	"github.com/stretchr/testify/require"
)

func TestBusiness_Intro(t *testing.T) {
	t.Skip()

	tSuite := testsuite.New(t)
	tSuite.UseLiveBackendAppDB()
	tSuite.FakeAPI().App() // init fakes

	bizRegistration := tSuite.FakeContainer().GetBizRegistration()
	// d0517ba8-3fd2-4c41-ba19-88cb96da7983
	userID := "43ebe70c-b147-47f5-ac11-262cfb8952bb"
	userDetails, err := bizRegistration.UserDetails(context.Background(), userID, &registration.QueryFilterUser{
		ID: null.StringFrom(userID),
	})
	require.NoError(t, err)
	fmt.Println("===== userDetails:", strutil.GetAsJson(userDetails))
}

func TestBusiness_UserCallState(t *testing.T) {
	t.Skip()

	tSuite := testsuite.New(t)
	tSuite.UseLiveBackendAppDB()
	tSuite.FakeAPI().App() // init fakes

	bizRegistration := tSuite.FakeContainer().GetBizRegistration()
	ucs, err := bizRegistration.UserCallState(context.Background(), &registration.User{
		ID: "8b4fc2de-f5e6-4848-875e-343ccc9c0049",
	})
	require.NoError(t, err)
	fmt.Println("===== ucs:", strutil.GetAsJson(ucs))
}
