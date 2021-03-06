// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package report

import (
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestManagerSuite is a test suite for the report manager.
type TestManagerSuite struct {
	suite.Suite

	m      Manager
	rpUUID string
}

// TestManager is an entry of suite TestManagerSuite.
func TestManager(t *testing.T) {
	suite.Run(t, &TestManagerSuite{})
}

// SetupSuite prepares test env for suite TestManagerSuite.
func (suite *TestManagerSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()

	suite.m = NewManager()
}

// SetupTest prepares env for test cases.
func (suite *TestManagerSuite) SetupTest() {
	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
	}

	uuid, err := suite.m.Create(rp)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), uuid)
	suite.rpUUID = uuid
}

// TearDownTest clears test env for test cases.
func (suite *TestManagerSuite) TearDownTest() {
	// No delete method defined in manager as no requirement,
	// so, to clear env, call dao method here
	err := scan.DeleteReport(suite.rpUUID)
	require.NoError(suite.T(), err)
}

// TestManagerCreateWithExisting tests the case that a copy already is there when creating report.
func (suite *TestManagerSuite) TestManagerCreateWithExisting() {
	err := suite.m.UpdateStatus(suite.rpUUID, job.SuccessStatus.String(), 2000)
	require.NoError(suite.T(), err)

	rp := &scan.Report{
		Digest:           "d1000",
		RegistrationUUID: "ruuid",
		MimeType:         v1.MimeTypeNativeReport,
	}

	uuid, err := suite.m.Create(rp)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), uuid)

	assert.NotEqual(suite.T(), suite.rpUUID, uuid)
	suite.rpUUID = uuid
}

// TestManagerGetBy tests the get by method.
func (suite *TestManagerSuite) TestManagerGetBy() {
	l, err := suite.m.GetBy("d1000", "ruuid", []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))
	assert.Equal(suite.T(), suite.rpUUID, l[0].UUID)

	l, err = suite.m.GetBy("d1000", "ruuid", nil)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))
	assert.Equal(suite.T(), suite.rpUUID, l[0].UUID)

	l, err = suite.m.GetBy("d1000", "", nil)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))
	assert.Equal(suite.T(), suite.rpUUID, l[0].UUID)
}

// TestManagerUpdateJobID tests update job ID method.
func (suite *TestManagerSuite) TestManagerUpdateJobID() {
	l, err := suite.m.GetBy("d1000", "ruuid", []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	oldJID := l[0].JobID

	err = suite.m.UpdateScanJobID(suite.rpUUID, "jID1001")
	require.NoError(suite.T(), err)

	l, err = suite.m.GetBy("d1000", "ruuid", []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	assert.NotEqual(suite.T(), oldJID, l[0].JobID)
	assert.Equal(suite.T(), "jID1001", l[0].JobID)
}

// TestManagerUpdateStatus tests update status method
func (suite *TestManagerSuite) TestManagerUpdateStatus() {
	l, err := suite.m.GetBy("d1000", "ruuid", []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	oldSt := l[0].Status

	err = suite.m.UpdateStatus(suite.rpUUID, job.SuccessStatus.String(), 10000)
	require.NoError(suite.T(), err)

	l, err = suite.m.GetBy("d1000", "ruuid", []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	assert.NotEqual(suite.T(), oldSt, l[0].Status)
	assert.Equal(suite.T(), job.SuccessStatus.String(), l[0].Status)
}

// TestManagerUpdateReportData tests update job report data.
func (suite *TestManagerSuite) TestManagerUpdateReportData() {
	err := suite.m.UpdateReportData(suite.rpUUID, "{\"a\":1000}", 1000)
	require.NoError(suite.T(), err)

	l, err := suite.m.GetBy("d1000", "ruuid", []string{v1.MimeTypeNativeReport})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 1, len(l))

	assert.Equal(suite.T(), "{\"a\":1000}", l[0].Report)
}
