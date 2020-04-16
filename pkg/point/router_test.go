package point

import (
	"fmt"
	"os"
	"testing"
	"time"

	"encoding/json"
	"net/http"

	"github.com/readr-media/readr-restful-point/config"
	"github.com/readr-media/readr-restful-point/internal/rrsql"
	tc "github.com/readr-media/readr-restful-point/internal/test"
)

type mockPointsAPI struct{}

type testInterface interface {
	setup(in interface{})
	teardown()
}

var mockPointsDS = []PointsProject{
	PointsProject{
		Points: Points{
			PointsID:   1,
			MemberID:   1,
			ObjectType: 1,
			ObjectID:   1,
			Points:     500,
			CreatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 1, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  rrsql.NullInt{Int: 1, Valid: true},
			UpdatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 2, 17, 17, 0, 0, time.UTC), Valid: true}},
	},
	PointsProject{
		Points: Points{
			PointsID:   2,
			MemberID:   1,
			ObjectType: 2,
			ObjectID:   3,
			Points:     300,
			CreatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 1, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  rrsql.NullInt{Int: 0, Valid: true},
			UpdatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 2, 17, 17, 0, 0, time.UTC), Valid: true}},
	},
	PointsProject{
		Points: Points{
			PointsID:   3,
			MemberID:   2,
			ObjectType: 1,
			ObjectID:   1,
			Points:     500,
			CreatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 1, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  rrsql.NullInt{Int: 1, Valid: true},
			UpdatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 2, 17, 17, 0, 0, time.UTC), Valid: true}},
	},
	PointsProject{
		Points: Points{
			PointsID:   4,
			MemberID:   1,
			ObjectType: 2,
			ObjectID:   23,
			Points:     100,
			CreatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 2, 17, 15, 0, 0, time.UTC), Valid: true},
			UpdatedBy:  rrsql.NullInt{Int: 1, Valid: true},
			UpdatedAt:  rrsql.NullTime{Time: time.Date(2018, 3, 5, 17, 17, 0, 0, time.UTC), Valid: true}},
	},
}

func (a *mockPointsAPI) Get(args *PointsArgs) (result []PointsProject, err error) {
	for _, v := range mockPointsDS {

		// Check member_id
		if v.MemberID == args.ID {
			// Check Object Type
			if args.ObjectType != nil {
				if v.ObjectType == int(*args.ObjectType) {
					// Check if there are object_id filter
					if args.ObjectIDs != nil {
						for _, o := range args.ObjectIDs {
							if v.ObjectID == o {
								result = append(result, v)
							}
						}
					} else {
						result = append(result, v)
					}
				}
			} else {
				result = append(result, v)
			}
		}
	}
	return result, err
}

func (a *mockPointsAPI) Insert(pts PointsToken) (result int, id int, err error) {

	args := PointsArgs{ID: pts.MemberID}
	if total, err := a.Get(&args); err == nil {
		for _, v := range total {
			result += int(v.Points.Points)
		}
		// mockPointsDS = append(a.mockPointsDS, PointsProject{Points: pts, Title: rrsql.NullString{"", false}})
		result -= pts.Points.Points
	}
	return result, 1, err
}

func TestMain(m *testing.M) {

	_, err := config.LoadConfig("../../config/main.json")
	if err != nil {
		panic(fmt.Errorf("Invalid application configuration: %s", err))
	}

	tc.SetRoutes(&Router)

	PointsAPI = new(mockPointsAPI)

	os.Exit(m.Run())
}

type TestStep struct {
	init     func()
	teardown func()
	register testInterface
	name     string
	cases    []tc.GenericTestcase
}

func TestRoutePoints(t *testing.T) {

	//Only Test if query parameter change would result in correct points history number change
	asserter := func(resp string, testcase tc.GenericTestcase, t *testing.T) {
		var Response = struct {
			Items []PointsProject `json:"_items"`
		}{}
		err := json.Unmarshal([]byte(resp), &Response)
		if err != nil {
			t.Errorf("%s, Unexpected result body: %v\n", testcase.Name, resp)
		}
		if len(Response.Items) != testcase.Resp.(int) {
			t.Errorf("%s, expect points history length to be %v but get %d", testcase.Name, testcase.Resp, len(Response.Items))
		}
	}
	t.Run("Get", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"BothTypePoints", "GET", `/points/1`, ``, http.StatusOK, 3},
			tc.GenericTestcase{"SingleTypePoints", "GET", `/points/1/2`, ``, http.StatusOK, 2},
			tc.GenericTestcase{"WithObjectID", "GET", `/points/1/2?object_ids=[23]`, ``, http.StatusOK, 1},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})
	t.Run("Insert", func(t *testing.T) {
		for _, testcase := range []tc.GenericTestcase{
			tc.GenericTestcase{"Deprecated Object Type: project", "POST", `/points`, `{"member_id":1,"object_type": 1}`, http.StatusBadRequest, `{"Error":"ObjectType Deprecated"}`},
			tc.GenericTestcase{"Deprecated Object Type: topup", "POST", `/points`, `{"member_id":1,"object_type": 3}`, http.StatusBadRequest, `{"Error":"ObjectType Deprecated"}`},
			tc.GenericTestcase{"Invalid Currency Value", "POST", `/points`, `{"member_id":1,"currency": -100,"object_type":5}`, http.StatusBadRequest, `{"Error":"Invalid Payment Amount"}`},
			tc.GenericTestcase{"Invalid ObjectType For Currency", "POST", `/points`, `{"member_id":1,"object_type": 4,"currency": 100}`, http.StatusBadRequest, `{"Error":"Currency Not Supported By ObjectType"}`},
			tc.GenericTestcase{"Missing Payment Token", "POST", `/points`, `{"member_id":1,"object_type": 5,"currency": 100}`, http.StatusBadRequest, `{"Error":"Invalid Token"}`},
			tc.GenericTestcase{"InvalidMemberInfo", "POST", `/points`, `{"member_id":1,"object_type": 5,"currency": 100, "token": "token"}`, http.StatusBadRequest, `{"Error":"Invalid Payment Info"}`},
			tc.GenericTestcase{"InvalidObjectID", "POST", `/points`, `{"member_id":1,"object_type": 2,"points": 100}`, http.StatusBadRequest, `{"Error":"Invalid Object ID"}`},
			tc.GenericTestcase{"InvalidObjectID", "POST", `/points`, `{"object_type": 2,"points": 100}`, http.StatusBadRequest, `{"Error":"Invalid ObjectType With Anonymous User"}`},

			tc.GenericTestcase{"Basic Project Memo", "POST", `/points`, `{"member_id":1,"object_type": 2,"object_id": 1,"currency": 50,"points": 50,"token":"token","member_name":"name","member_phone":"phone","member_mail":"mail"}`, http.StatusOK, `{"id":1,"points":850}`},
			tc.GenericTestcase{"Basic Gift", "POST", `/points`, `{"member_id":1,"object_type": 4,"object_id": 1,"points": -50, "reason": "System"}`, http.StatusOK, `{"id":1,"points":950}`},
			tc.GenericTestcase{"Basic Donate", "POST", `/points`, `{"member_id":1,"object_type": 5,"object_id": 1,"currency": 100,"token":"token","member_name":"name","member_phone":"phone","member_mail":"mail"}`, http.StatusOK, `{"id":1,"points":900}`},
		} {
			tc.GenericDoTest(testcase, t, asserter)
		}
	})
}
