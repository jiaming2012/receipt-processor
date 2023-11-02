package main

import (
	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"
	"jiaming2012/receipt-processor/utils"

	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type PurchaseItemDTO struct {
	Description string `yaml:"description"`
}

type PurchaseItemGroupDTO struct {
	Name          string            `yaml:"name"`
	PurchaseItems []PurchaseItemDTO `yaml:"purchase_items"`
}

func main() {
	// Get the PROJECT_DIR environment variable
	projectDir := os.Getenv("PROJECT_DIR")
	if projectDir == "" {
		log.Fatalf("PROJECT_DIR environment variable not set")
	}

	// Change the current working directory to the root directory.
	if err := os.Chdir(projectDir); err != nil {
		panic(err)
	}

	// Read the YAML file
	yamlFile, err := os.ReadFile("fixtures/purchase_item_groups.yaml")
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	// Parse the YAML file
	// var groups []PurchaseItemGroup
	var groups []PurchaseItemGroupDTO
	err = yaml.Unmarshal(yamlFile, &groups)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
	}

	// Connect to the database
	if err := database.Setup(); err != nil {
		panic(err)
	}

	utils.SetupDB()

	db := database.GetDB()
	defer database.ReleaseDB()

	// Truncase the purchase_item_groups table
	err = db.Exec("TRUNCATE TABLE purchase_item_groups CASCADE").Error
	if err != nil {
		log.Fatalf("Error truncating purchase_item_groups: %v", err)
	}

	// Truncate the purchase_item_group_purchase_items table
	err = db.Exec("TRUNCATE TABLE purchase_item_group_purchase_items CASCADE").Error
	if err != nil {
		log.Fatalf("Error truncating purchase_item_group_purchase_items: %v", err)
	}

	// Loop through the purchase item groups
	for _, group := range groups {
		purchaseItemGroup := models.PurchaseItemGroup{
			Name: group.Name,
		}

		if err := db.Create(&purchaseItemGroup).Error; err != nil {
			log.Fatalf("Error creating group: %v", err)
		}

		// Find all menu items whose description matches the purchase item group description
		var purchaseItems []models.PurchaseItem
		for _, purchaseItem := range group.PurchaseItems {
			err := db.Where("description LIKE ?", purchaseItem.Description).Find(&purchaseItems).Error
			if err != nil {
				log.Fatalf("Error finding menu items: %v", err)
			}

			for _, purchaseItem := range purchaseItems {
				// add the menu item to the purchase item group
				err = db.Exec("INSERT INTO purchase_item_group_purchase_items (purchase_item_group_id, purchase_item_id) VALUES ($1, $2)", purchaseItemGroup.ID, purchaseItem.ID).Error
				if err != nil {
					log.Fatalf("Error inserting into purchase_item_group_purchase_items: %v", err)
				}
			}

			log.Infof("Inserted %d purchase items into purchase_item_group %s, using purchase item predicate %s", len(purchaseItems), purchaseItemGroup.Name, purchaseItem.Description)
		}
	}
}
