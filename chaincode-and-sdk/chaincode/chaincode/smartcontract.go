package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type Failure struct {
	ID             string  `json:"ID"`
	Description    string  `json:"description"`
	Price          float64 `json:"price"`
	Repaired       bool    `json:"repaired"`
}

type Car struct {
	ID             string `json:"ID"`
	Brand          string `json:"brand"`
	Model          string `json:"model"`
	Year           int    `json:"year"`
	Color          string `json:"color"`
	OwnerID        string `json:"ownerID"`
	Failures       []Failure `json:"failures"`
	Price          float64 `json:"price"`
}

type Person struct {
	ID             string  `json:"ID"`
	Name           string  `json:"name"`
	Surname        string  `json:"surname"`
	Email          string  `json:"email"`
	Balance        float64 `json:"balance"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	persons := []Person{
		{ID: "1", Name: "Pero", Surname: "Peric", Email: "pero.peric@maildrop.cc", Balance: 100.00},
		{ID: "2", Name: "Milos", Surname: "Vucic", Email: "milos.vucic@maildrop.cc", Balance: 150.00},
		{ID: "3", Name: "Edo", Surname: "Teka", Email: "edo.teka@maildrop.cc", Balance: 200.00},
	}
	
	cars := []Car{
		{ID: "101", Brand: "Mercedes", Model: "C", Year: 2021, Color: "siva", OwnerID: "1", Price: 150, Failures: make([]Failure, 0)},
		{ID: "102", Brand: "BMW", Model: "250", Year: 2019, Color: "crna", OwnerID: "1", Price: 150, Failures: make([]Failure, 0)},
		{ID: "103", Brand: "Toyota", Model: "Hybrid", Year: 2017, Color: "plava", OwnerID: "2", Price: 50, Failures: make([]Failure, 0)},
		{ID: "104", Brand: "Opel", Model: "Astra C", Year: 2010, Color: "siva", OwnerID: "2", Price: 50, Failures: make([]Failure, 0)},
		{ID: "105", Brand: "Audi", Model: "R8", Year: 2013, Color: "bela", OwnerID: "3", Price: 200, Failures: make([]Failure, 0)},
		{ID: "106", Brand: "Volkswagen", Model: "Golf 7", Year: 2015, Color: "crvena", OwnerID: "3", Price: 100, Failures: make([]Failure, 0)},
	}

	for _, person := range persons {
		personJSON, err := json.Marshal(person)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(person.ID, personJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	
	for _, car := range cars {
		carJSON, err := json.Marshal(car)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(car.ID, carJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// Change color
func (s *SmartContract) ChangeColor(ctx contractapi.TransactionContextInterface, id string, color string) error {
	carJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if carJSON == nil {
		return fmt.Errorf("the car %s does not exist", id)
	}

	var car Car
	err = json.Unmarshal(carJSON, &car)
	if err != nil {
		return err
	}
	car.Color = color

	carJSON, err = json.Marshal(car)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, carJSON)
}

// Repair failure
func (s *SmartContract) RepairFailure(ctx contractapi.TransactionContextInterface, id string, failure string) error {
	carJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if carJSON == nil {
		return fmt.Errorf("the car %s does not exist", id)
	}

	var car Car
	err = json.Unmarshal(carJSON, &car)
	if err != nil {
		return err
	}
	
	personJSON, err := ctx.GetStub().GetState(car.OwnerID)
    	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if personJSON == nil {
		return fmt.Errorf("the person %s does not exist", car.OwnerID)
	}
	var person Person
	err = json.Unmarshal(personJSON, &person)
	if err != nil {
		return err
	}
	
	for idx, failr := range car.Failures {
		if failr.ID == failure {
			car.Failures[idx].Repaired = true
			person.Balance = person.Balance - failr.Price
			break
		}
    	}

	carJSON, err = json.Marshal(car)
	if err != nil {
		return err
	}
	personJSON, err = json.Marshal(person)
	if err != nil {
		return err
	}

	ctx.GetStub().PutState(car.OwnerID, personJSON)

	return ctx.GetStub().PutState(id, carJSON)
}

// Transfer ownership
func (s *SmartContract) TransferOwnership(ctx contractapi.TransactionContextInterface, id string, acceptFailures bool, newOwner string) error {
	carJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if carJSON == nil {
		return fmt.Errorf("the car %s does not exist", id)
	}

	var car Car
	err = json.Unmarshal(carJSON, &car)
	if err != nil {
		return err
	}
	
	if car.OwnerID == newOwner {
    		return fmt.Errorf("this is your car, you can't buy it")
    	}
	
	var failureExists bool
	for _, failr := range car.Failures {
		if !failr.Repaired {
			failureExists = true
			break
		}
    	}
    	
    	if failureExists && !acceptFailures {
    		return fmt.Errorf("car has failures")
    	}
    	
    	// Get old owner
    	personJSONOwner, err := ctx.GetStub().GetState(car.OwnerID)
    	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if personJSONOwner == nil {
		return fmt.Errorf("the person %s does not exist", car.OwnerID)
	}
	var personOwner Person
	err = json.Unmarshal(personJSONOwner, &personOwner)
	if err != nil {
		return err
	}
	
    	// Get new owner
    	personJSON, err := ctx.GetStub().GetState(newOwner)
    	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if personJSON == nil {
		return fmt.Errorf("the person %s does not exist", newOwner)
	}
	var person Person
	err = json.Unmarshal(personJSON, &person)
	if err != nil {
		return err
	}
    	
	// check failures
    	if !failureExists {
    		if person.Balance >= car.Price {
    			car.OwnerID = newOwner
    			person.Balance = person.Balance - car.Price
    			personOwner.Balance = personOwner.Balance + car.Price
    		} else {
    			return fmt.Errorf("the person %s does not have enough money", newOwner)
    		}

    	} else {
    		if acceptFailures {
    			var adjustedPrice float64 = car.Price
    			for _, failr := range car.Failures {
				if !failr.Repaired {
					adjustedPrice = adjustedPrice - failr.Price
				}
    			}
    			if person.Balance >= adjustedPrice {
    				car.OwnerID = newOwner
    				person.Balance = person.Balance - adjustedPrice
    				personOwner.Balance = personOwner.Balance + adjustedPrice
	    		} else {
		    		return fmt.Errorf("the person %s does not have enough money", newOwner)
	    		}
    		}
    	}

	carJSON, err = json.Marshal(car)
	if err != nil {
		return err
	}
	personJSON, err = json.Marshal(person)
	if err != nil {
		return err
	}
	personJSONOwner, err = json.Marshal(personOwner)
	if err != nil {
		return err
	}

	ctx.GetStub().PutState(newOwner, personJSON)
	ctx.GetStub().PutState(personOwner.ID, personJSONOwner)
	return ctx.GetStub().PutState(id, carJSON)
}

// Create failure
func (s *SmartContract) CreateFailure(ctx contractapi.TransactionContextInterface, id string, failureID string, desc string, price float64) error {
	carJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if carJSON == nil {
		return fmt.Errorf("the car %s does not exist", id)
	}

	var car Car
	err = json.Unmarshal(carJSON, &car)
	if err != nil {
		return err
	}

	var totalPrice float64
	for _, failr := range car.Failures {
		if !failr.Repaired {
			totalPrice = totalPrice + failr.Price
		}
	}
	totalPrice = totalPrice + price
	
	if totalPrice >= car.Price {
		return ctx.GetStub().DelState(id)
	}
	
	// price is ok
	failure := Failure{
		ID:             failureID,
		Description:    desc,
		Price:          price,
		Repaired:       false,
	}
	car.Failures = append(car.Failures, failure)
	
	// update car
	carJSON, err = json.Marshal(car)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, carJSON)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// Get all cars by color
func (s *SmartContract) GetAllCarsByColor(ctx contractapi.TransactionContextInterface, color string) ([]*Car, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var cars []*Car
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var car Car
		err = json.Unmarshal(queryResponse.Value, &car)
		if err != nil {
			continue
		}
		if car.Color == color {
			cars = append(cars, &car)
		}
	}

	return cars, nil
}


// Get all cars by color and owner
// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllCarsByColorAndOwner(ctx contractapi.TransactionContextInterface, color string, owner string) ([]*Car, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var cars []*Car
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var car Car
		err = json.Unmarshal(queryResponse.Value, &car)
		if err != nil {
			continue
		}
		if car.Color == color && car.OwnerID == owner {
			cars = append(cars, &car)
		}
	}

	return cars, nil
}
