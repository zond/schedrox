package crm

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"monotone/se.oort.schedrox/common"
	"monotone/se.oort.schedrox/domain"
	"monotone/se.oort.schedrox/search"
	"reflect"
	"sort"
	"strings"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
)

const (
	contactRootName = "Contacts"
)

type contactPage struct {
	Contacts Contacts
	Total    int
}

func ContactKeyForId(k *datastore.Key) string {
	return fmt.Sprintf("Contact{Id:%v}", k)
}

func contactsKeyForPrefix(prefix string) string {
	return fmt.Sprintf("Contacts{Prefix:%v}", prefix)
}

func contactsKeyForOffsetAndLimit(offset, limit int) string {
	return fmt.Sprintf("Contacts{Offset:%v,Limit:%v}", offset, limit)
}

func contactsKeyForDomain(domain *datastore.Key) string {
	return fmt.Sprintf("Contacts{Domain:%v}", domain)
}

type Contacts []Contact

func (self Contacts) Len() int {
	return len(self)
}
func (self Contacts) Less(i, j int) bool {
	return bytes.Compare([]byte(self[i].Name), []byte(self[j].Name)) < 0
}
func (self Contacts) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type Contact struct {
	Id                  *datastore.Key `json:"id" datastore:"-"`
	Name                string         `json:"name"`
	ContactFamilyName   string         `json:"contact_family_name"`
	ContactGivenName    string         `json:"contact_given_name"`
	MobilePhone         string         `json:"mobile_phone"`
	Email               string         `json:"email"`
	AddressLine1        string         `json:"address_line_1"`
	AddressLine2        string         `json:"address_line_2"`
	AddressLine3        string         `json:"address_line_3"`
	BillingAddressLine1 string         `json:"billing_address_line_1"`
	BillingAddressLine2 string         `json:"billing_address_line_2"`
	BillingAddressLine3 string         `json:"billing_address_line_3"`
	Information         string         `json:"information"`
	Reference           string         `json:"reference"`
	OrganizationNumber  string         `json:"organization_number"`
	GravatarHash        string         `json:"gravatar_hash" datastore:"-"`
	EmailBounceBytes    []byte         `json:"-"`
	EmailBounce         string         `json:"email_bounce" datastore:"-"`
}

func (self *Contact) SendMail(c gaecontext.HTTPContext, subject, body, extra_bcc string) {
	replyTo := fmt.Sprintf("noreply@%v.appspotmail.com", appengine.AppID(c))
	dom := domain.GetDomain(c, self.Id.Parent())
	if dom.FromAddress != "" {
		replyTo = dom.FromAddress
	}
	var bcc []string
	if extra_bcc != "" {
		bcc = []string{extra_bcc}
	}
	msg := &mail.Message{
		// Karbin ville inte ha den här funktionen, från-adressen var för ful.
		//		Sender:  fmt.Sprintf("%v <cb+%v@%v.appspotmail.com>", common.SenderUser(c), common.EncodeBase64(fmt.Sprintf("%v:%v", dom.Name, self.Name)), appengine.AppID(c)),
		Sender:  replyTo,
		ReplyTo: replyTo,
		To:      []string{self.Email},
		Subject: subject,
		Body:    body,
		Bcc:     bcc,
	}
	if appengine.AppID(c) != "schedev" {
		if err := mail.Send(c, msg); err != nil {
			buf := new(bytes.Buffer)
			fmt.Fprintf(buf, "%v while sending %v", err, string(buf.Bytes()))
			self.EmailBounce = string(buf.Bytes())
			self.Save(c, dom.Id)
		}
	} else {
		log.Debugf(c, "Would have sent\n%v\nif appengine.AppID(c) [%+v] == schedrox", msg, appengine.AppID(c))
	}
	if appengine.IsDevAppServer() {
		log.Debugf(c, "Body in above mail: %v", body)
	}
}

func (self *Contact) process(c gaecontext.HTTPContext) *Contact {
	self.EmailBounce = string(self.EmailBounceBytes)
	m := md5.New()
	io.WriteString(m, strings.ToLower(strings.TrimSpace(self.Email)))
	self.GravatarHash = fmt.Sprintf("%x", m.Sum(nil))
	return self
}

func (self *Contact) CopyFrom(n *Contact) *Contact {
	self.ContactFamilyName = n.ContactFamilyName
	self.ContactGivenName = n.ContactGivenName
	self.MobilePhone = n.MobilePhone
	self.Name = n.Name
	self.Email = n.Email
	self.Information = n.Information
	self.Reference = n.Reference
	self.OrganizationNumber = n.OrganizationNumber
	self.BillingAddressLine1 = n.BillingAddressLine1
	self.BillingAddressLine2 = n.BillingAddressLine2
	self.BillingAddressLine3 = n.BillingAddressLine3
	self.AddressLine1 = n.AddressLine1
	self.AddressLine2 = n.AddressLine2
	self.AddressLine3 = n.AddressLine3
	self.EmailBounce = n.EmailBounce
	return self
}

func DeleteContact(c gaecontext.HTTPContext, key, dom *datastore.Key) {
	if !dom.Equal(key.Parent()) {
		panic(fmt.Errorf("%v is not parent of %v", dom, key))
	}
	if err := c.Transaction(func(c gaecontext.HTTPContext) (err error) {
		if err = datastore.Delete(c, key); err != nil {
			return
		}
		search.Deindex(c, key)
		common.MemDel(c, contactsKeyForDomain(dom))
		common.MemDel(c, ContactKeyForId(key))
		return nil
	}, false); err != nil {
		panic(err)
	}
}

func findContact(c gaecontext.HTTPContext, key *datastore.Key) (result *Contact) {
	var contact Contact
	err := datastore.Get(c, key, &contact)
	if err == nil {
		result = &contact
		result.Id = key
	}
	return
}

func GetContact(c gaecontext.HTTPContext, k *datastore.Key, dom *datastore.Key) *Contact {
	if !k.Parent().Equal(dom) {
		panic(fmt.Errorf("%v is not parent of %v", dom, k))
	}
	var contact Contact
	if common.Memoize(c, ContactKeyForId(k), &contact, func() interface{} {
		return findContact(c, k)
	}) {
		return (&contact).process(c)
	}
	return nil
}

func (self *Contact) Save(c gaecontext.HTTPContext, dom *datastore.Key) *Contact {
	if err := c.Transaction(func(c gaecontext.HTTPContext) (err error) {
		self.EmailBounceBytes = []byte(self.EmailBounce)
		if self.Id == nil {
			self.Id, err = datastore.Put(c, datastore.NewKey(c, "Contact", "", 0, dom), self)
		} else {
			_, err = datastore.Put(c, self.Id, self)
		}
		if err != nil {
			return
		}
		search.Deindex(c, self.Id)
		search.IndexString(c, self.Id, "default", strings.Join([]string{self.Name, self.ContactFamilyName, self.ContactGivenName, self.Email, self.MobilePhone, self.AddressLine1, self.AddressLine2, self.AddressLine3, self.BillingAddressLine1, self.BillingAddressLine2, self.BillingAddressLine3, self.Reference, self.OrganizationNumber, self.Information}, " "))
		common.MemDel(c, contactsKeyForDomain(dom))
		common.MemDel(c, ContactKeyForId(self.Id))
		return nil
	}, false); err != nil {
		panic(err)
	}
	return self
}

func findContactsByFieldAndPrefix(c gaecontext.HTTPContext, dom *datastore.Key, badPrefix string, max int, badField string, funnel chan Contact, done chan bool) chan bool {
	prefix := strings.ToLower(badPrefix)
	field := fmt.Sprintf("%vDown", badField)
	result := make(chan bool)
	go func() {
		var contacts Contacts
		ids, err := datastore.NewQuery("Contact").Ancestor(dom).Filter(fmt.Sprintf("%v>", field), prefix).Order(field).Limit(max).GetAll(c, &contacts)
		if err != nil {
			panic(err)
		}
		for index, contact := range contacts {
			if strings.Index(reflect.ValueOf(contact).FieldByName(field).String(), prefix) == 0 {
				contact.Id = ids[index]
				funnel <- contact
			}
		}
		if done != nil {
			<-done
		}
		result <- true
	}()
	return result
}

func findContactsByPrefix(c gaecontext.HTTPContext, dom *datastore.Key, prefix string) (result Contacts) {
	ids, err := search.Search(c, dom, "Contact", "default", prefix, true, 128, &result)
	if err != nil {
		panic(err)
	}
	for index, id := range ids {
		result[index].Id = id
	}
	return
}

func GetContactsByPrefix(c gaecontext.HTTPContext, dom *datastore.Key, prefix string) (result Contacts) {
	common.Memoize2(c, contactsKeyForDomain(dom), contactsKeyForPrefix(prefix), &result, func() interface{} {
		return findContactsByPrefix(c, dom, prefix)
	})
	sort.Sort(result)
	for index, _ := range result {
		(&result[index]).process(c)
	}
	if result == nil {
		result = make(Contacts, 0)
	}
	return
}

func findContactsByDomain(c gaecontext.HTTPContext, d *datastore.Key, offset, limit int) (result contactPage) {
	var err error
	query := datastore.NewQuery("Contact").Ancestor(d).Order("Name")
	result.Total, err = query.Count(c)
	if err != nil {
		panic(err)
	}
	ids, err := query.Offset(offset).Limit(limit).GetAll(c, &result.Contacts)
	if err != nil {
		panic(err)
	}
	for index, id := range ids {
		result.Contacts[index].Id = id
	}
	return
}

func GetContactsByDomain(c gaecontext.HTTPContext, d *datastore.Key, offset, limit int) (Contacts, int) {
	var page contactPage
	common.Memoize2(c, contactsKeyForDomain(d), contactsKeyForOffsetAndLimit(offset, limit), &page, func() interface{} {
		return findContactsByDomain(c, d, offset, limit)
	})
	for index, _ := range page.Contacts {
		(&page.Contacts[index]).process(c)
	}
	return page.Contacts, page.Total
}
