package sumologic

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceSumologicContent() *schema.Resource {
	return &schema.Resource{
		Create: resourceSumologicContentCreate,
		Read:   resourceSumologicContentRead,
		Delete: resourceSumologicContentDelete,

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"config": {
				Type:             schema.TypeString,
				ValidateFunc:     validation.StringIsJSON,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
		},
	}
}

func resourceSumologicContentRead(d *schema.ResourceData, meta interface{}) error {
	log.Println("====Begin Content Read====")

	c := meta.(*Client)
	//retrieve the content Id from the state
	id := d.Id()
	log.Printf("Search for Content Id: %s", id)

	log.Println("Looking up content...")
	content, err := c.GetContent(id)

	//Error retrieving content
	if err != nil {
		return err
	}

	if content == nil {
		log.Printf("[WARN] Content not found, removing from state: %v - %v", id, err)
		d.SetId("")
		return nil
	}

	log.Println("Read Values:")
	log.Printf("ParentId: %s", content.ParentId)
	log.Printf("Config: %s", content.Config)
	log.Printf("Name: %s", content.Name)

	// Write the newly read content object into the schema
	d.Set("config", content.Config)

	log.Println("====End Content Read====")
	return nil
}

func resourceSumologicContentDelete(d *schema.ResourceData, meta interface{}) error {
	log.Println("====Begin Content Delete====")
	log.Printf("Deleting Content Id: %s", d.Id())
	c := meta.(*Client)
	log.Println("====End Content Delete====")
	return c.DeleteContent(d.Id())
}

func resourceSumologicContentCreate(d *schema.ResourceData, meta interface{}) error {
	log.Println("====Begin Content Create====")
	c := meta.(*Client)

	//If there is no id in the state, then we need to create the object
	if d.Id() == "" {

		//Load all the data we have from the schema into a Content Struct
		content := resourceToContent(d)
		log.Println("Newly populated content values:")
		log.Printf("ParentId: %s", content.ParentId)
		log.Printf("Config: %s", content.Config)

		//Call create content with our newly populated struct
		id, err := c.CreateContent(*content)

		//Error during CreateContent
		if err != nil {
			return err
		}

		log.Println("Saving Id to state...")
		d.SetId(id)
		log.Printf("ContentId: %s", id)
		log.Printf("ContentType: %s", content.Type)

	}

	log.Println("====End Content Create====")

	//After creating an object, we read it again to make sure the state is properly saved
	return resourceSumologicContentRead(d, meta)
}

func resourceToContent(d *schema.ResourceData) *Content {
	log.Println("Loading data from schema to Content struct...")
	var content Content

	_ = json.Unmarshal([]byte(d.Get("config").(string)), &content)

	content.Children = []Content{}
	content.ParentId = d.Get("parent_id").(string)
	content.Config = d.Get("config").(string)
	content.ID = d.Id()

	return &content
}
