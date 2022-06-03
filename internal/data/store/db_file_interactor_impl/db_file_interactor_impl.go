package db_file_interactor_impl

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/k0marov/golang-auth/internal/data/models"
	"github.com/k0marov/golang-auth/internal/domain/entities"
)

type DBFileInteractorImpl struct {
	dbFileName string
}

func NewDBFileInteractor(dbFileName string) *DBFileInteractorImpl {
	return &DBFileInteractorImpl{
		dbFileName: dbFileName,
	}
}

func (d *DBFileInteractorImpl) ReadUsers() ([]models.UserModel, error) {
	dbFile, err := os.OpenFile(d.dbFileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return []models.UserModel{}, fmt.Errorf("error opening file while reading users: %w", err)
	}
	defer dbFile.Close()
	csvReader := csv.NewReader(dbFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		return []models.UserModel{}, fmt.Errorf("got an error while reading users: %w", err)
	}

	users := []models.UserModel{}
	for _, record := range records {
		user, err := sliceToUserModel(record)
		if err != nil {
			return []models.UserModel{}, fmt.Errorf("error converting csv row to user model: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (d *DBFileInteractorImpl) WriteUser(newUser models.UserModel) error {
	dbFile, err := os.OpenFile(d.dbFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening file writing appending user: %w", err)
	}
	defer dbFile.Close()
	csvWriter := csv.NewWriter(dbFile)

	record := []string{
		strconv.Itoa(newUser.Id),
		newUser.Username,
		newUser.StoredPass,
		newUser.AuthToken.Token,
	}
	err = csvWriter.Write(record)
	if err != nil {
		return fmt.Errorf("error writing record to csv: %w", err)
	}
	csvWriter.Flush()
	if err = csvWriter.Error(); err != nil {
		return fmt.Errorf("error flushing a record to csv: %w", err)
	}
	return nil
}

const numberOfModelFields = 4

func sliceToUserModel(slice []string) (models.UserModel, error) {
	if len(slice) != numberOfModelFields {
		return models.UserModel{}, fmt.Errorf("incorrect amount of columns in a csv row: %v", slice)
	}
	id, err := strconv.Atoi(slice[0])
	if err != nil {
		return models.UserModel{}, fmt.Errorf("error converting id to int: %w", err)
	}
	return models.UserModel{
		Id:         id,
		Username:   slice[1],
		StoredPass: slice[2],
		AuthToken:  entities.Token{Token: slice[3]},
	}, nil
}
