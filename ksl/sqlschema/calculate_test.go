package sqlschema_test

import (
	_ "ksl/postgres"
	"ksl/schema"
	"ksl/sqlschema"
	"testing"
)

func TestCalculate(t *testing.T) {
	data := `
model Album {
	id       int     @id @map("AlbumId")
	title    string  @default("TestDefaultTitle")
	artist_id int
	tracks   Track[]
	artist   Artist  @ref(fields: [artist_id], references: [id])
}

model Track {
	id            int           @id @map("TrackId")
	name          string
	composer      string?
	milliseconds  int
	unit_price    float
	album_id      int?
	genre_id      int?
	media_type_id int
	media_type    MediaType     @ref(fields: [media_type_id], references: [id])
	genre         Genre?        @ref(fields: [genre_id], references: [id])
	album         Album?        @ref(fields: [album_id], references: [id])
	invoice_lines InvoiceLine[]
}

model MediaType {
	id    int     @id @map("MediaTypeId")
	name  string?
	track Track[]
}

model Genre {
	id     int     @id @map("GenreId")
	name   string?
	tracks Track[]
}

model Artist {
	id     int     @id @map("ArtistId")
	name   string?
	albums Album[]
}

model Customer {
	id             int       @id @map("CustomerId")
	first_name     string
	last_name      string
	company        string?
	address        string?
	city           string?
	state          State?
	country        string?
	postal_code    string?
	phone          string?
	fax            string?
	email          string
	support_rep_id int?
	support_rep    Employee? @ref(fields: [support_rep_id], references: [id])
	invoices       Invoice[]
}

model Employee {
	id          int        @id @map("EmployeeId")
	first_name  string
	last_name   string
	title       string?
	birth_date  datetime?
	hire_date   datetime?
	address    	string?
	city       	string?
	state      	State?
	country    	string?
	postal_code string?
	phone      	string?
	fax        	string?
	email      	string?
	customers  	Customer[]
}

enum State {
	CA
	NY
	TX
}

model Invoice {
	id                  int           @id @map("InvoiceId")
	invoice_date        datetime
	billing_address     string?
	billing_city        string?
	billing_state       string?
	billing_country     string?
	billing_postal_code string?
	total               float
	customer_id         int
	customer            Customer      @ref(fields: [customer_id], references: [id])
	lines               InvoiceLine[]
}

model InvoiceLine {
	id         int     @id @map("InvoiceLineId")
	unit_price float
	quantity   int
	invoice_id int
	track_id   int
	invoice    Invoice @ref(fields: [invoice_id], references: [id])
	track      Track   @ref(fields: [track_id], references: [id])
}

model Playlist {
	id   int     @id @map("PlaylistId")
	name string?
}`
	sch := schema.Parse([]byte(data), "test.ksl")
	db := sqlschema.CalculateSqlSchema(sch, "public")
	_ = db
}

func TestCalculateM2M(t *testing.T) {
	data := `
model Person {
	id     int     @id
	cars  Car[]
}

model Car {
	id     int     @id
	owners Person[]
}
`
	sch := schema.Parse([]byte(data), "test.ksl")
	db := sqlschema.CalculateSqlSchema(sch, "public")
	_ = db
}
