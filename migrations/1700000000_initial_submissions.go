package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection := core.NewBaseCollection("submissions")

		// Add text field
		collection.Fields.Add(
			&core.TextField{
				Name:     "text",
				Required: true,
			},
		)

		// Add notes field for admin comments
		collection.Fields.Add(
			&core.TextField{
				Name:     "notes",
				Required: false,
			},
		)

		// Add status field with options
		collection.Fields.Add(
			&core.SelectField{
				Name:      "status",
				Required:  true,
				MaxSelect: 1,
				Values:    []string{"new", "approved", "hidden", "deleted", "done"},
			},
		)

		// Add timestamp fields
		collection.Fields.Add(
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
		)

		collection.Fields.Add(
			&core.AutodateField{
				Name:     "updated",
				OnUpdate: true,
			},
		)

		return app.Save(collection)
	}, func(app core.App) error {
		// Revert: delete the collection
		collection, err := app.FindCollectionByNameOrId("submissions")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
