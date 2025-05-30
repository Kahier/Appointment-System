package api

import (
	"backend/helper"
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Hospital struct {
	HospitalCode int       `bson:"hospitalCode" json:"hospitalCode"`
	HospitalName string    `bson:"hospitalName" json:"hospitalName"`
	DistrictCode int       `bson:"districtCode" json:"districtCode"`
	ProvinceCode int       `bson:"provinceCode" json:"provinceCode"`
	Fields       []int     `bson:"fields" json:"fields"`
	CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time `bson:"updatedAt" json:"updatedAt"`
}

func GetAllHospitals(client *mongo.Client) []Hospital {
	collection := client.Database("healthcare").Collection("hospitals")

	filter := bson.D{{}}

	cursor, _ := collection.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())

	var hospitals []Hospital
	cursor.All(context.TODO(), &hospitals)

	return hospitals
}

func GetHospital(client *mongo.Client, hospitalCode int) (*Hospital, error) {
	collection := client.Database("healthcare").Collection("hospitals")

	filter := bson.D{{Key: "hospitalCode", Value: hospitalCode}}
	var hospital Hospital

	err := collection.FindOne(context.TODO(), filter).Decode(&hospital)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("no such doctor")
		}
		return nil, err
	}

	return &hospital, nil
}

func GetHospitalsByProvince(client *mongo.Client, provinceCode int) []Hospital {

	collection := client.Database("healthcare").Collection("hospitals")

	filter := bson.D{{Key: "provinceCode", Value: provinceCode}}

	cursor, _ := collection.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())

	var hospitals []Hospital
	for cursor.Next(context.TODO()) {
		var hospital Hospital
		if err := cursor.Decode(&hospital); err != nil {
			log.Fatal(err)
		}
		hospitals = append(hospitals, hospital)
	}

	return hospitals
}

func GetHospitalsByDistrict(client *mongo.Client, districtCode int) []Hospital {

	collection := client.Database("healthcare").Collection("hospitals")

	filter := bson.D{{Key: "districtCode", Value: districtCode}}

	cursor, _ := collection.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())

	var hospitals []Hospital
	for cursor.Next(context.TODO()) {
		var hospital Hospital
		if err := cursor.Decode(&hospital); err != nil {
			log.Fatal(err)
		}
		hospitals = append(hospitals, hospital)
	}

	return hospitals
}

func DeleteHospital(client *mongo.Client, hospitalCode int) {
	collection := client.Database("healthcare").Collection("hospitals")
	collection.FindOneAndDelete(context.TODO(), bson.D{{Key: "hospitalCode", Value: hospitalCode}})
	doctors, err := GetDoctorsByHospitalCode(client, hospitalCode)
	if err != nil {
		return
	}
	for _, doctor := range doctors {
		DeleteDoctor(client, doctor.DoctorCode)
	}

}

func CreateHospital(client *mongo.Client, hospital Hospital) {
	collection := client.Database("healthcare").Collection("hospitals")
	hospital.HospitalCode = helper.GenerateIntID(5)
	hospital.CreatedAt = time.Now()
	hospital.UpdatedAt = time.Now()
	collection.InsertOne(context.TODO(), hospital)
}

func UpdateHospital(client *mongo.Client, hospital Hospital) {
	collection := client.Database("healthcare").Collection("hospitals")
	hospital.UpdatedAt = time.Now()

	_, err := collection.ReplaceOne(
		context.TODO(),
		bson.M{"hospitalCode": hospital.HospitalCode},
		hospital,
	)

	if err != nil {
		log.Println("Error updating hospital:", err)
	}
}
