package main

import (
	"errors"
	"os"

	"github.com/jackc/pgconn"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"jiaming2012/receipt-processor/database"
	"jiaming2012/receipt-processor/models"
)

type PurchaseItemDTO struct {
	Description string   `yaml:"description"`
	Exclusions  []string `yaml:"exclusions"`
}

type PurchaseItemGroupDTO struct {
	Name          string            `yaml:"name"`
	Tags          []string          `yaml:"tags"`
	PurchaseItems []PurchaseItemDTO `yaml:"purchase_items"`
}

func setupDB() {
	log.Info("Setting up database ...")
	if err := database.Setup(); err != nil {
		log.Errorf("failed to setup database: %v", err)
		return
	}
	db := database.GetDB()
	defer database.ReleaseDB()

	db.AutoMigrate(&models.ToastOrderDetail{})
	db.AutoMigrate(&models.Store{})
	db.AutoMigrate(&models.MetaV2{})
	db.AutoMigrate(&models.PurchaseV2{})
	db.AutoMigrate(&models.PurchaseItem{})
	db.AutoMigrate(&models.ToastItemSelectionDetail{})
	db.AutoMigrate(&models.MenuItem{})
	db.AutoMigrate(&models.PurchaseItemGroup{})
	db.AutoMigrate(&models.PurchaseItem{})
	db.AutoMigrate(&models.Tag{})

	log.Info("Db setup complete!")
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
	var groups []PurchaseItemGroupDTO
	err = yaml.Unmarshal(yamlFile, &groups)
	if err != nil {
		log.Fatalf("Error parsing YAML file: %v", err)
	}

	// Connect to the database
	if err := database.Setup(); err != nil {
		panic(err)
	}

	setupDB()

	db := database.GetDB()
	defer database.ReleaseDB()

	// Truncate the tags table
	err = db.Exec("TRUNCATE TABLE tags CASCADE").Error
	if err != nil {
		log.Fatalf("Error truncating tags: %v", err)
	}

	// Truncast the purchase_item_groups table
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

		// Create or fetch tags
		var tags []*models.Tag
		for _, tag := range group.Tags {
			var existingTag models.Tag
			err := db.Where("name = ?", tag).First(&existingTag).Error
			if err != nil {
				existingTag = models.Tag{Name: tag}
				if err := db.Create(&existingTag).Error; err != nil {
					log.Fatalf("Error creating tag: %v", err)
				}
			}

			tags = append(tags, &existingTag)
		}

		// Add the tags to the purchase item group
		purchaseItemGroup.Tags = tags

		if err := db.Create(&purchaseItemGroup).Error; err != nil {
			log.Fatalf("Error creating group: %v", err)
		}

		// Find all menu items whose description matches the purchase item group description
		var purchaseItems []models.PurchaseItem
		for _, purchaseItem := range group.PurchaseItems {
			// Fetch all allExclusions
			var allExclusions []models.PurchaseItem
			for _, exclusion := range purchaseItem.Exclusions {
				var exclusions []models.PurchaseItem
				err := db.Where("description LIKE ?", exclusion).Find(&exclusions).Error
				if err != nil {
					log.Fatalf("Error finding menu items: %v", err)
				}

				allExclusions = append(allExclusions, exclusions...)
			}

			err = db.Where("description LIKE ?", purchaseItem.Description).Find(&purchaseItems).Error
			if err != nil {
				log.Fatalf("Error finding menu items: %v", err)
			}

			for _, purchaseItem := range purchaseItems {
				for _, exclusion := range allExclusions {
					if purchaseItem.ID == exclusion.ID {
						log.Infof("Excluding purchase item %s from purchase item group %s", purchaseItem.Description, purchaseItemGroup.Name)
						continue
					}
				}

				// add the menu item to the purchase item group
				err = db.Exec("INSERT INTO purchase_item_group_purchase_items (purchase_item_group_id, purchase_item_id) VALUES ($1, $2)", purchaseItemGroup.ID, purchaseItem.ID).Error
				if err != nil {
					// conditionally check and cast the error to pgconn.PgError to check the error code
					var pgErr *pgconn.PgError
					if errors.As(err, &pgErr); pgErr != nil && pgErr.Code == "23505" { // pq: duplicate key value violates unique constraint. Occurs when link was already set.
						log.Infof("Purchase item %s already exists in purchase item group %s", purchaseItem.Description, purchaseItemGroup.Name)
						continue
					}

					log.Fatalf("Error inserting into purchase_item_group_purchase_items [%s, %s]: %v", purchaseItemGroup.Name, purchaseItem.Description, err)
				}
			}

			log.Infof("Inserted %d purchase items into purchase_item_group %s, using purchase item predicate %s", len(purchaseItems), purchaseItemGroup.Name, purchaseItem.Description)
		}
	}
}
