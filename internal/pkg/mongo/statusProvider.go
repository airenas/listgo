package mongo

import (
	"context"

	"github.com/airenas/listgo/internal/app/status/api"
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/airenas/listgo/internal/pkg/err"
	"github.com/airenas/listgo/internal/pkg/persistence"
	"github.com/airenas/listgo/internal/pkg/progress"
	"github.com/airenas/listgo/internal/pkg/status"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
)

// StatusProvider provides transcription status from mongo db
type StatusProvider struct {
	SessionProvider *SessionProvider
}

// NewStatusProvider creates StatusProvider instance
func NewStatusProvider(sessionProvider *SessionProvider) (*StatusProvider, error) {
	f := StatusProvider{SessionProvider: sessionProvider}
	return &f, nil
}

// Get retrieves status from DB
func (fs StatusProvider) Get(id string) (*api.TranscriptionResult, error) {
	cmdapp.Log.Infof("Retrieving status %s", id)

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := fs.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(store).Collection(statusTable)

	var m persistence.Status
	err = c.FindOne(ctx, bson.M{"ID": id}).Decode(&m)
	if err == mgo.ErrNoDocuments {
		cmdapp.Log.Infof("ID not found %s", id)
		return newNotFoundResult(id), nil
	}

	if err != nil {
		return nil, err
	}

	result := api.TranscriptionResult{ID: id}

	result.Status = m.Status
	result.ErrorCode = m.ErrorCode
	result.Error = m.Error
	stv := status.From(result.Status)
	result.Progress = progress.Convert(stv)
	if stv == status.Completed {
		result.RecognizedText, err = getResultText(ctx, session, id)
	}
	result.AudioReady = m.AudioReady
	result.AvailableResults = m.AvailableResults

	return &result, err
}

// Get retrieves status from DB
func getResultText(ctx context.Context, session mgo.Session, id string) (string, error) {
	cmdapp.Log.Infof("Retrieving result %s", id)

	c := session.Client().Database(store).Collection(resultTable)

	var m persistence.Result

	err := c.FindOne(ctx, bson.M{"ID": id}).Decode(&m)
	if err == mgo.ErrNoDocuments {
		cmdapp.Log.Infof("ID not found %s", id)
		return "", nil
	}
	if err != nil {
		return "", errors.Wrap(err, "can't load results")
	}
	return m.Text, nil
}

func newNotFoundResult(ID string) *api.TranscriptionResult {
	result := api.TranscriptionResult{ID: ID}
	result.Status = "NOT_FOUND"
	result.ErrorCode = err.NotFoundCode
	result.Error = "Nežinomas ID: " + ID
	return &result
}
