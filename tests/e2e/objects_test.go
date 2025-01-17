package e2e

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	pubnub "github.com/pubnub/go"
	"github.com/stretchr/testify/assert"
)

func ActivateWithPAM() *pubnub.PubNub {
	pn := pubnub.NewPubNub(pamConfigCopy())
	return pn
}

func RunGrant(pn *pubnub.PubNub, users, spaces []string, read, write, manage, del, create, createPattern bool) []string {
	u := map[string]pubnub.UserSpacePermissions{}
	for _, user := range users {
		u[user] = pubnub.UserSpacePermissions{
			Read:   read,
			Write:  write,
			Manage: manage,
			Delete: del,
			Create: create,
		}
	}

	up := map[string]pubnub.UserSpacePermissions{}
	if createPattern && len(u) > 0 {
		up["^.*"] = pubnub.UserSpacePermissions{
			Read:   read,
			Write:  write,
			Manage: manage,
			Delete: del,
			Create: create,
		}
	}

	s := map[string]pubnub.UserSpacePermissions{}
	for _, space := range spaces {
		s[space] = pubnub.UserSpacePermissions{
			Read:   read,
			Write:  write,
			Manage: manage,
			Delete: del,
			Create: create,
		}
	}

	sp := map[string]pubnub.UserSpacePermissions{}
	if createPattern && len(s) > 0 {
		sp["^.*"] = pubnub.UserSpacePermissions{
			Read:   read,
			Write:  write,
			Manage: manage,
			Delete: del,
			Create: create,
		}

	}

	res, _, err := pn.GrantToken().TTL(3).
		//Channels(ch).
		//ChannelGroups(cg).
		Users(u).
		Spaces(s).
		Execute()
	fmt.Println(res)
	fmt.Println(err)

	token2 := ""

	if len(up) > 0 || len(sp) > 0 {
		res2, _, err2 := pn.GrantToken().TTL(3).
			//Channels(ch).
			//ChannelGroups(cg).
			UsersPattern(up).
			SpacesPattern(sp).
			Execute()
		fmt.Println(res2)
		fmt.Println(err2)
		token2 = res2.Data.Token
	}

	if res != nil {
		return []string{res.Data.Token, token2}
	}
	return []string{}
}

func SetPN(pn, pn2 *pubnub.PubNub, tokens []string) {
	pn.Config.SubscribeKey = pn2.Config.SubscribeKey
	pn.Config.PublishKey = pn2.Config.PublishKey
	pn.Config.SecretKey = ""
	pn.Config.Origin = pn2.Config.Origin
	pn.Config.Secure = pn2.Config.Secure

	pn.SetTokens(tokens)
	fmt.Println("========")
	fmt.Println(pn.GetTokens())
	fmt.Println("========")
}

func TestObjectsCreateUpdateGetDeleteUser(t *testing.T) {
	ObjectsCreateUpdateGetDeleteUserCommon(t, false, false)
}

func TestObjectsCreateUpdateGetDeleteUserWithPAM(t *testing.T) {
	ObjectsCreateUpdateGetDeleteUserCommon(t, true, false)
}

func TestObjectsCreateUpdateGetDeleteUserWithPAMWithoutSecKey(t *testing.T) {
	ObjectsCreateUpdateGetDeleteUserCommon(t, true, true)
}

func ObjectsCreateUpdateGetDeleteUserCommon(t *testing.T, withPAM, runWithoutSecretKey bool) {

	assert := assert.New(t)

	pn := pubnub.NewPubNub(configCopy())
	r := GenRandom()

	id := fmt.Sprintf("testuser_%d", r.Intn(99999))
	if withPAM {
		pn2 := ActivateWithPAM()
		if runWithoutSecretKey {
			tokens := RunGrant(pn2, []string{id}, []string{}, true, true, true, true, true, true)
			SetPN(pn, pn2, tokens)
		} else {
			pn = pn2
			RunGrant(pn, []string{id}, []string{}, true, true, false, true, true, false)
		}

	}

	name := "name"
	extid := "extid"
	purl := "profileurl"
	email := "email"

	custom := make(map[string]interface{})
	custom["a"] = "b"
	custom["c"] = "d"

	incl := []pubnub.PNUserSpaceInclude{
		pubnub.PNUserSpaceCustom,
	}

	res, st, err := pn.CreateUser().Include(incl).ID(id).Name(name).ExternalID(extid).ProfileURL(purl).Email(email).Custom(custom).Execute()
	assert.Nil(err)
	assert.Equal(200, st.StatusCode)
	assert.Equal(id, res.Data.ID)
	assert.Equal(name, res.Data.Name)
	assert.Equal(extid, res.Data.ExternalID)
	assert.Equal(purl, res.Data.ProfileURL)
	assert.Equal(email, res.Data.Email)
	assert.NotNil(res.Data.Created)
	assert.NotNil(res.Data.Updated)
	assert.NotNil(res.Data.ETag)
	assert.Equal("b", res.Data.Custom["a"])
	assert.Equal("d", res.Data.Custom["c"])

	email = "email2"

	res2, st2, err2 := pn.UpdateUser().Include(incl).ID(id).Name(name).ExternalID(extid).ProfileURL(purl).Email(email).Custom(custom).Execute()
	assert.Nil(err2)
	assert.Equal(200, st2.StatusCode)
	assert.Equal(id, res2.Data.ID)
	assert.Equal(name, res2.Data.Name)
	assert.Equal(extid, res2.Data.ExternalID)
	assert.Equal(purl, res2.Data.ProfileURL)
	assert.Equal(email, res2.Data.Email)
	assert.Equal(res.Data.Created, res2.Data.Created)
	assert.NotNil(res2.Data.Updated)
	assert.NotNil(res2.Data.ETag)
	assert.Equal("b", res2.Data.Custom["a"])
	assert.Equal("d", res2.Data.Custom["c"])

	res3, st3, err3 := pn.GetUser().Include(incl).ID(id).Execute()
	assert.Nil(err3)
	assert.Equal(200, st3.StatusCode)
	assert.Equal(id, res3.Data.ID)
	assert.Equal(name, res3.Data.Name)
	assert.Equal(extid, res3.Data.ExternalID)
	assert.Equal(purl, res3.Data.ProfileURL)
	assert.Equal(email, res3.Data.Email)
	assert.Equal(res.Data.Created, res3.Data.Created)
	assert.Equal(res2.Data.Updated, res3.Data.Updated)
	assert.Equal(res2.Data.ETag, res3.Data.ETag)
	assert.Equal("b", res3.Data.Custom["a"])
	assert.Equal("d", res3.Data.Custom["c"])

	//getusers
	res6, st6, err6 := pn.GetUsers().Include(incl).Limit(100).Count(true).Execute()
	assert.Nil(err6)
	assert.Equal(200, st6.StatusCode)
	assert.True(res6.TotalCount > 0)
	found := false
	for i := range res6.Data {
		if res6.Data[i].ID == id {
			assert.Equal(name, res6.Data[i].Name)
			assert.Equal(extid, res6.Data[i].ExternalID)
			assert.Equal(purl, res6.Data[i].ProfileURL)
			assert.Equal(email, res6.Data[i].Email)
			assert.Equal(res.Data.Created, res6.Data[i].Created)
			assert.Equal(res2.Data.Updated, res6.Data[i].Updated)
			assert.Equal(res2.Data.ETag, res6.Data[i].ETag)
			assert.Equal("b", res6.Data[i].Custom["a"])
			assert.Equal("d", res6.Data[i].Custom["c"])
			found = true
		}
	}
	assert.True(found)

	//delete
	res5, st5, err5 := pn.DeleteUser().ID(id).Execute()
	assert.Nil(err5)
	assert.Equal(200, st5.StatusCode)
	assert.Nil(res5.Data)

	//getuser
	res4, st4, err4 := pn.GetUser().Include(incl).ID(id).Execute()
	assert.NotNil(err4)
	assert.Nil(res4)
	assert.Equal(404, st4.StatusCode)

}

func TestObjectsCreateUpdateGetDeleteSpace(t *testing.T) {
	ObjectsCreateUpdateGetDeleteSpaceCommon(t, false, false)
}

func TestObjectsCreateUpdateGetDeleteSpaceWithPAM(t *testing.T) {
	ObjectsCreateUpdateGetDeleteSpaceCommon(t, true, false)
}

func TestObjectsCreateUpdateGetDeleteSpaceWithPAMWithoutSecKey(t *testing.T) {
	ObjectsCreateUpdateGetDeleteSpaceCommon(t, true, true)
}

func ObjectsCreateUpdateGetDeleteSpaceCommon(t *testing.T, withPAM, runWithoutSecretKey bool) {
	assert := assert.New(t)

	pn := pubnub.NewPubNub(configCopy())

	r := GenRandom()

	id := fmt.Sprintf("testspace_%d", r.Intn(99999))

	if withPAM {
		pn2 := ActivateWithPAM()
		if runWithoutSecretKey {
			tokens := RunGrant(pn2, []string{}, []string{id}, true, true, false, true, true, true)
			SetPN(pn, pn2, tokens)
		} else {
			pn = pn2
			RunGrant(pn, []string{}, []string{id}, true, true, false, true, true, false)
		}

	}
	name := "name"
	desc := "desc"

	custom := make(map[string]interface{})
	custom["a"] = "b"
	custom["c"] = "d"

	incl := []pubnub.PNUserSpaceInclude{
		pubnub.PNUserSpaceCustom,
	}

	res, st, err := pn.CreateSpace().Include(incl).ID(id).Name(name).Description(desc).Custom(custom).Execute()
	assert.Nil(err)
	assert.Equal(200, st.StatusCode)
	assert.Equal(id, res.Data.ID)
	assert.Equal(name, res.Data.Name)
	assert.Equal(desc, res.Data.Description)
	assert.NotNil(res.Data.Created)
	assert.NotNil(res.Data.Updated)
	assert.NotNil(res.Data.ETag)
	assert.Equal("b", res.Data.Custom["a"])
	assert.Equal("d", res.Data.Custom["c"])

	desc = "desc2"

	res2, st2, err2 := pn.UpdateSpace().Include(incl).ID(id).Name(name).Description(desc).Custom(custom).Execute()
	assert.Nil(err2)
	assert.Equal(200, st2.StatusCode)
	assert.Equal(id, res2.Data.ID)
	assert.Equal(name, res2.Data.Name)
	assert.Equal(desc, res2.Data.Description)
	assert.Equal(res.Data.Created, res2.Data.Created)
	assert.NotNil(res2.Data.Updated)
	assert.NotNil(res2.Data.ETag)
	assert.Equal("b", res2.Data.Custom["a"])
	assert.Equal("d", res2.Data.Custom["c"])

	res3, st3, err3 := pn.GetSpace().Include(incl).ID(id).Execute()
	assert.Nil(err3)
	assert.Equal(200, st3.StatusCode)
	assert.Equal(id, res3.Data.ID)
	assert.Equal(name, res3.Data.Name)
	assert.Equal(desc, res3.Data.Description)
	assert.Equal(res.Data.Created, res3.Data.Created)
	assert.Equal(res2.Data.Updated, res3.Data.Updated)
	assert.Equal(res2.Data.ETag, res3.Data.ETag)
	assert.Equal("b", res3.Data.Custom["a"])
	assert.Equal("d", res3.Data.Custom["c"])

	//getusers
	res6, st6, err6 := pn.GetSpaces().Include(incl).Limit(100).Count(true).Execute()
	assert.Nil(err6)
	assert.Equal(200, st6.StatusCode)
	assert.True(res6.TotalCount > 0)
	found := false
	for i := range res6.Data {
		if res6.Data[i].ID == id {
			assert.Equal(name, res6.Data[i].Name)
			assert.Equal(desc, res6.Data[i].Description)
			assert.Equal(res.Data.Created, res6.Data[i].Created)
			assert.Equal(res2.Data.Updated, res6.Data[i].Updated)
			assert.Equal(res2.Data.ETag, res6.Data[i].ETag)
			assert.Equal("b", res6.Data[i].Custom["a"])
			assert.Equal("d", res6.Data[i].Custom["c"])
			found = true
		}
	}
	assert.True(found)

	//delete
	res5, st5, err5 := pn.DeleteSpace().ID(id).Execute()
	assert.Nil(err5)
	assert.Equal(200, st5.StatusCode)
	assert.Nil(res5.Data)

	//getuser
	res4, st4, err4 := pn.GetSpace().Include(incl).ID(id).Execute()
	assert.NotNil(err4)
	assert.Nil(res4)
	assert.Equal(404, st4.StatusCode)

}

func TestObjectsMemberships(t *testing.T) {
	ObjectsMembershipsCommon(t, false, false)
}

func TestObjectsMembershipsWithPAM(t *testing.T) {
	ObjectsMembershipsCommon(t, true, false)
}

// PASSES after adding PAM checks for Update Members
func TestObjectsMembershipsWithPAMWithoutSecKey(t *testing.T) {
	ObjectsMembershipsCommon(t, true, true)
}

func ObjectsMembershipsCommon(t *testing.T, withPAM, runWithoutSecretKey bool) {
	assert := assert.New(t)

	limit := 100
	count := true

	pn := pubnub.NewPubNub(configCopy())

	r := GenRandom()

	userid := fmt.Sprintf("testuser_%d", r.Intn(99999))
	spaceid := fmt.Sprintf("testspace_%d", r.Intn(99999))

	if withPAM {
		pn2 := ActivateWithPAM()
		if runWithoutSecretKey {
			tokens := RunGrant(pn2, []string{userid}, []string{spaceid}, true, true, true, true, true, true)
			SetPN(pn, pn2, tokens)
		} else {
			pn = pn2
			RunGrant(pn, []string{userid}, []string{spaceid}, true, true, true, true, true, false)
		}

	}

	name := "name"
	extid := "extid"
	purl := "profileurl"
	email := "email"

	custom := make(map[string]interface{})
	custom["a"] = "b"
	custom["c"] = "d"

	incl := []pubnub.PNUserSpaceInclude{
		pubnub.PNUserSpaceCustom,
	}

	res, st, err := pn.CreateUser().Include(incl).ID(userid).Name(name).ExternalID(extid).ProfileURL(purl).Email(email).Custom(custom).Execute()
	assert.Nil(err)
	assert.Equal(200, st.StatusCode)
	assert.Equal(userid, res.Data.ID)
	assert.Equal(name, res.Data.Name)
	assert.Equal(extid, res.Data.ExternalID)
	assert.Equal(purl, res.Data.ProfileURL)
	assert.Equal(email, res.Data.Email)
	assert.NotNil(res.Data.Created)
	assert.NotNil(res.Data.Updated)
	assert.NotNil(res.Data.ETag)
	assert.Equal("b", res.Data.Custom["a"])
	assert.Equal("d", res.Data.Custom["c"])

	desc := "desc"
	custom2 := make(map[string]interface{})
	custom2["a1"] = "b1"
	custom2["c1"] = "d1"

	res2, st2, err2 := pn.CreateSpace().Include(incl).ID(spaceid).Name(name).Description(desc).Custom(custom2).Execute()
	assert.Nil(err2)
	assert.Equal(200, st2.StatusCode)
	assert.Equal(spaceid, res2.Data.ID)
	assert.Equal(name, res2.Data.Name)
	assert.Equal(desc, res2.Data.Description)
	assert.NotNil(res2.Data.Created)
	assert.NotNil(res2.Data.Updated)
	assert.NotNil(res2.Data.ETag)
	assert.Equal("b1", res2.Data.Custom["a1"])
	assert.Equal("d1", res2.Data.Custom["c1"])

	userid2 := fmt.Sprintf("testuser_%d", r.Intn(99999))

	_, st3, err3 := pn.CreateUser().Include(incl).ID(userid2).Name(name).ExternalID(extid).ProfileURL(purl).Email(email).Custom(custom).Execute()
	assert.Nil(err3)
	assert.Equal(200, st3.StatusCode)

	spaceid2 := fmt.Sprintf("testspace_%d", r.Intn(99999))

	_, st4, err4 := pn.CreateSpace().Include(incl).ID(spaceid2).Name(name).Description(desc).Custom(custom2).Execute()
	assert.Nil(err4)
	assert.Equal(200, st4.StatusCode)

	inclSm := []pubnub.PNMembersInclude{
		pubnub.PNMembersCustom,
		pubnub.PNMembersUser,
		pubnub.PNMembersUserCustom,
	}

	custom3 := make(map[string]interface{})
	custom3["a3"] = "b3"
	custom3["c3"] = "d3"

	in := pubnub.PNMembersInput{
		ID:     userid,
		Custom: custom3,
	}

	inArr := []pubnub.PNMembersInput{
		in,
	}

	//Add Space Memberships

	resAdd, stAdd, errAdd := pn.ManageMembers().SpaceID(spaceid).Add(inArr).Update([]pubnub.PNMembersInput{}).Remove([]pubnub.PNMembersRemove{}).Include(inclSm).Limit(limit).Count(count).Execute()
	assert.Nil(errAdd)
	assert.Equal(200, stAdd.StatusCode)
	if errAdd == nil {

		found := false
		assert.True(resAdd.TotalCount > 0)
		fmt.Println("resAdd-->", resAdd)

		for i := range resAdd.Data {
			if resAdd.Data[i].ID == userid {
				found = true
				assert.Equal(custom3["a3"], resAdd.Data[i].Custom["a3"])
				assert.Equal(custom3["c3"], resAdd.Data[i].Custom["c3"])
				assert.Equal(userid, resAdd.Data[0].User.ID)
				assert.Equal(name, resAdd.Data[0].User.Name)
				assert.Equal(extid, resAdd.Data[0].User.ExternalID)
				assert.Equal(purl, resAdd.Data[0].User.ProfileURL)
				assert.Equal(email, resAdd.Data[0].User.Email)
				assert.Equal(custom["a"], resAdd.Data[0].User.Custom["a"])
				assert.Equal(custom["c"], resAdd.Data[0].User.Custom["c"])
			}
		}
		assert.True(found)
	} else {
		fmt.Println("ManageMembers->", errAdd.Error())
	}

	//Update Space Memberships
	if !withPAM {

		custom4 := make(map[string]interface{})
		custom4["a2"] = "b2"
		custom4["c2"] = "d2"

		up := pubnub.PNMembersInput{
			ID:     userid,
			Custom: custom4,
		}

		upArr := []pubnub.PNMembersInput{
			up,
		}

		resUp, stUp, errUp := pn.ManageMembers().SpaceID(spaceid).Add([]pubnub.PNMembersInput{}).Update(upArr).Remove([]pubnub.PNMembersRemove{}).Include(inclSm).Limit(limit).Count(count).Execute()
		assert.Nil(errUp)
		assert.Equal(200, stUp.StatusCode)
		if errUp == nil {
			assert.True(resUp.TotalCount > 0)
			foundUp := false
			for i := range resUp.Data {
				if resUp.Data[i].ID == userid {
					foundUp = true
					assert.Equal("b2", resUp.Data[i].Custom["a2"])
					assert.Equal("d2", resUp.Data[i].Custom["c2"])
					assert.Equal(userid, resAdd.Data[0].User.ID)
					assert.Equal(name, resAdd.Data[0].User.Name)
					assert.Equal(extid, resAdd.Data[0].User.ExternalID)
					assert.Equal(purl, resAdd.Data[0].User.ProfileURL)
					assert.Equal(email, resAdd.Data[0].User.Email)
					assert.Equal(custom["a"], resAdd.Data[0].User.Custom["a"])
					assert.Equal(custom["c"], resAdd.Data[0].User.Custom["c"])

				}
			}
			assert.True(foundUp)
		} else {
			fmt.Println("ManageMembers->", errUp.Error())
		}
	}
	//Get Space Memberships

	inclMemberships := []pubnub.PNMembershipsInclude{
		pubnub.PNMembershipsCustom,
		pubnub.PNMembershipsSpace,
		pubnub.PNMembershipsSpaceCustom,
	}

	fmt.Println("GetMemberships ====>")

	resGetMem, stGetMem, errGetMem := pn.GetMemberships().UserID(userid).Include(inclMemberships).Limit(limit).Count(count).Execute()
	foundGetMem := false
	assert.Nil(errGetMem)
	if errGetMem == nil {
		for i := range resGetMem.Data {
			if resGetMem.Data[i].ID == spaceid {
				foundGetMem = true
				assert.Equal(name, resGetMem.Data[i].Space.Name)
				assert.Equal(desc, resGetMem.Data[i].Space.Description)
				assert.Equal("b1", resGetMem.Data[i].Space.Custom["a1"])
				assert.Equal("d1", resGetMem.Data[i].Space.Custom["c1"])
				if withPAM {
					assert.Equal("b3", resGetMem.Data[i].Custom["a3"])
					assert.Equal("d3", resGetMem.Data[i].Custom["c3"])
				} else {
					assert.Equal("b2", resGetMem.Data[i].Custom["a2"])
					assert.Equal("d2", resGetMem.Data[i].Custom["c2"])
				}
			}
		}
		assert.Equal(200, stGetMem.StatusCode)
		assert.True(foundGetMem)
	} else {
		fmt.Println("GetMemberships->", errGetMem.Error())
	}

	//Remove Space Memberships
	re := pubnub.PNMembersRemove{
		ID: userid,
	}

	reArr := []pubnub.PNMembersRemove{
		re,
	}
	resRem, stRem, errRem := pn.ManageMembers().SpaceID(spaceid).Add([]pubnub.PNMembersInput{}).Update([]pubnub.PNMembersInput{}).Remove(reArr).Include(inclSm).Limit(limit).Count(count).Execute()
	assert.Nil(errRem)
	assert.Equal(200, stRem.StatusCode)
	if errRem == nil {

		foundRem := false
		for i := range resRem.Data {
			if resRem.Data[i].ID == userid {
				foundRem = true
				assert.Equal("b2", resRem.Data[i].Custom["a2"])
				assert.Equal("d2", resRem.Data[i].Custom["c2"])
				assert.Equal(userid, resRem.Data[0].User.ID)
				assert.Equal(name, resRem.Data[0].User.Name)
				assert.Equal(extid, resRem.Data[0].User.ExternalID)
				assert.Equal(purl, resRem.Data[0].User.ProfileURL)
				assert.Equal(email, resRem.Data[0].User.Email)
				assert.Equal(custom["a"], resRem.Data[0].User.Custom["a"])
				assert.Equal(custom["c"], resRem.Data[0].User.Custom["c"])

			}
		}
		assert.False(foundRem)
	} else {
		fmt.Println("ManageMembers->", errRem.Error())
	}

	inMem := pubnub.PNMembershipsInput{
		ID:     spaceid2,
		Custom: custom3,
	}

	inArrMem := []pubnub.PNMembershipsInput{
		inMem,
	}

	//Add user memberships
	resManageMemAdd, stManageMemAdd, errManageMemAdd := pn.ManageMemberships().UserID(userid2).Add(inArrMem).Update([]pubnub.PNMembershipsInput{}).Remove([]pubnub.PNMembershipsRemove{}).Include(inclMemberships).Limit(limit).Count(count).Execute()
	fmt.Println("resManageMemAdd -->", resManageMemAdd)
	assert.Nil(errManageMemAdd)
	assert.Equal(200, stManageMemAdd.StatusCode)
	if errManageMemAdd == nil {
		foundManageMembers := false
		for i := range resManageMemAdd.Data {
			if resManageMemAdd.Data[i].ID == spaceid2 {
				assert.Equal(spaceid2, resManageMemAdd.Data[i].Space.ID)
				assert.Equal(name, resManageMemAdd.Data[i].Space.Name)
				assert.Equal(desc, resManageMemAdd.Data[i].Space.Description)
				assert.Equal(custom2["a1"], resManageMemAdd.Data[i].Space.Custom["a1"])
				assert.Equal(custom2["c1"], resManageMemAdd.Data[i].Space.Custom["c1"])
				assert.Equal(custom3["a3"], resManageMemAdd.Data[i].Custom["a3"])
				assert.Equal(custom3["c3"], resManageMemAdd.Data[i].Custom["c3"])
				foundManageMembers = true
			}
		}
		assert.True(foundManageMembers)
	} else {
		fmt.Println("ManageMemberships->", errManageMemAdd.Error())
	}

	// //Update user memberships

	custom5 := make(map[string]interface{})
	custom5["a5"] = "b5"
	custom5["c5"] = "d5"

	upMem := pubnub.PNMembershipsInput{
		ID:     spaceid2,
		Custom: custom5,
	}

	upArrMem := []pubnub.PNMembershipsInput{
		upMem,
	}

	resManageMemUp, stManageMemUp, errManageMemUp := pn.ManageMemberships().UserID(userid2).Add([]pubnub.PNMembershipsInput{}).Update(upArrMem).Remove([]pubnub.PNMembershipsRemove{}).Include(inclMemberships).Limit(limit).Count(count).Execute()
	fmt.Println("resManageMemUp -->", resManageMemUp)
	assert.Nil(errManageMemUp)
	assert.Equal(200, stManageMemUp.StatusCode)
	if errManageMemUp == nil {
		foundManageMembersUp := false
		for i := range resManageMemUp.Data {
			if resManageMemUp.Data[i].ID == spaceid2 {
				assert.Equal(spaceid2, resManageMemUp.Data[i].Space.ID)
				assert.Equal(name, resManageMemUp.Data[i].Space.Name)
				assert.Equal(desc, resManageMemUp.Data[i].Space.Description)
				assert.Equal(custom2["a1"], resManageMemAdd.Data[i].Space.Custom["a1"])
				assert.Equal(custom2["c1"], resManageMemAdd.Data[i].Space.Custom["c1"])
				assert.Equal(custom5["a5"], resManageMemUp.Data[i].Custom["a5"])
				assert.Equal(custom5["c5"], resManageMemUp.Data[i].Custom["c5"])
				foundManageMembersUp = true
			}
		}
		assert.True(foundManageMembersUp)
	} else {
		fmt.Println("ManageMemberships->", errManageMemUp.Error())
	}

	// //Get members
	resGetMembers, stGetMembers, errGetMembers := pn.GetMembers().SpaceID(spaceid2).Include(inclSm).Limit(limit).Count(count).Execute()
	fmt.Println("resGetMembers -->", resGetMembers)
	assert.Nil(errGetMembers)
	assert.Equal(200, stGetMembers.StatusCode)
	if errGetMembers == nil {
		foundGetMembers := false

		for i := range resGetMembers.Data {
			if resGetMembers.Data[i].ID == userid2 {
				foundGetMembers = true
				assert.Equal(name, resGetMembers.Data[i].User.Name)
				assert.Equal(extid, resGetMembers.Data[i].User.ExternalID)
				assert.Equal(purl, resGetMembers.Data[i].User.ProfileURL)
				assert.Equal(email, resGetMembers.Data[i].User.Email)
				assert.Equal(custom["a"], resGetMembers.Data[i].User.Custom["a"])
				assert.Equal(custom["c"], resGetMembers.Data[i].User.Custom["c"])
				assert.Equal(custom5["a5"], resGetMembers.Data[i].Custom["a5"])
				assert.Equal(custom5["c5"], resGetMembers.Data[i].Custom["c5"])
			}
		}
		assert.True(foundGetMembers)
	} else {
		fmt.Println("GetMembers->", errGetMembers.Error())
	}

	// //Remove user memberships

	reMem := pubnub.PNMembershipsRemove{
		ID: spaceid2,
	}

	reArrMem := []pubnub.PNMembershipsRemove{
		reMem,
	}
	resManageMemRem, stManageMemRem, errManageMemRem := pn.ManageMemberships().UserID(userid2).Add([]pubnub.PNMembershipsInput{}).Update([]pubnub.PNMembershipsInput{}).Remove(reArrMem).Include(inclMemberships).Limit(limit).Count(count).Execute()
	assert.Nil(errManageMemRem)
	assert.Equal(200, stManageMemRem.StatusCode)
	if errManageMemRem == nil {

		foundManageMemRem := false
		for i := range resManageMemRem.Data {
			if resManageMemRem.Data[i].ID == spaceid2 {
				foundManageMemRem = true
			}
		}
		assert.False(foundManageMemRem)
	} else {
		fmt.Println("ManageMemberships->", errManageMemRem.Error())
	}

	//delete
	res5, st5, err5 := pn.DeleteUser().ID(userid).Execute()
	assert.Nil(err5)
	assert.Equal(200, st5.StatusCode)

	assert.Nil(res5.Data)

	//delete
	res6, st6, err6 := pn.DeleteSpace().ID(spaceid).Execute()
	assert.Nil(err6)
	assert.Equal(200, st6.StatusCode)
	assert.Nil(res6.Data)

	//delete
	res52, st52, err52 := pn.DeleteUser().ID(userid2).Execute()
	assert.Nil(err52)
	assert.Equal(200, st52.StatusCode)
	if res52 != nil {
		assert.Nil(res52.Data)
	}

	//delete
	res62, st62, err62 := pn.DeleteSpace().ID(spaceid2).Execute()
	assert.Nil(err62)
	assert.Equal(200, st62.StatusCode)
	if res62 != nil {
		assert.Nil(res62.Data)
	}

}

func TestObjectsListeners(t *testing.T) {
	ObjectsListenersCommon(t, false, false)
}

func TestObjectsListenersWithPAM(t *testing.T) {
	ObjectsListenersCommon(t, true, false)
}

func TestObjectsListenersWithPAMWithoutSecKey(t *testing.T) {
	ObjectsListenersCommon(t, true, true)
}

func ObjectsListenersCommon(t *testing.T, withPAM, runWithoutSecretKey bool) {
	//Create channel names for Space and User
	eventWaitTime := 2
	assert := assert.New(t)

	limit := 100
	count := true

	pn := pubnub.NewPubNub(configCopy())
	pnSub := pubnub.NewPubNub(configCopy())

	r := GenRandom()

	userid := fmt.Sprintf("testlistuser_%d", r.Intn(99999))
	spaceid := fmt.Sprintf("testlistspace_%d", r.Intn(99999))
	if withPAM {
		pn2 := ActivateWithPAM()
		if runWithoutSecretKey {
			pn2.Config.Log = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
			tokens := RunGrant(pn2, []string{userid}, []string{spaceid}, true, true, true, true, true, true)
			SetPN(pn, pn2, tokens)
			SetPN(pnSub, pn2, tokens)
			//You have to use Grant v2 to subscribe
			pnSub.Config.AuthKey = "authKey"
			pn2.Grant().
				Read(true).Write(true).Manage(true).
				Channels([]string{userid, spaceid}).
				AuthKeys([]string{pnSub.Config.AuthKey}).
				Execute()
		} else {
			pn = pn2
			pnSub = pn2
			pn.Config.Log = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
			RunGrant(pn, []string{userid}, []string{spaceid}, true, true, true, true, true, false)
		}
	}
	pn.Config.Log = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	pnSub.Config.Log = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	//Subscribe to the channel names

	listener := pubnub.NewListener()

	var mut sync.RWMutex

	addUserToSpace := false
	addUserToSpace2 := false
	updateUserMem := false
	updateUser := false
	updateSpace := false
	removeUserFromSpace := false
	deleteUser := false
	deleteSpace := false

	doneConnected := make(chan bool)

	go func() {
		for {
			fmt.Println("Running =--->")
			select {

			case status := <-listener.Status:
				switch status.Category {
				case pubnub.PNConnectedCategory:
					doneConnected <- true
				default:
					fmt.Println(" --- status: ", status)
				}

			case userEvent := <-listener.UserEvent:

				fmt.Println(" --- UserEvent: ")
				fmt.Println(fmt.Sprintf("%s", userEvent))
				fmt.Println(fmt.Sprintf("userEvent.Channel: %s", userEvent.Channel))
				fmt.Println(fmt.Sprintf("userEvent.SubscribedChannel: %s", userEvent.SubscribedChannel))
				fmt.Println(fmt.Sprintf("userEvent.Event: %s", userEvent.Event))
				fmt.Println(fmt.Sprintf("userEvent.UserID: %s", userEvent.UserID))
				fmt.Println(fmt.Sprintf("userEvent.Description: %s", userEvent.Description))
				fmt.Println(fmt.Sprintf("userEvent.Timestamp: %s", userEvent.Timestamp))
				fmt.Println(fmt.Sprintf("userEvent.Name: %s", userEvent.Name))
				fmt.Println(fmt.Sprintf("userEvent.ExternalID: %s", userEvent.ExternalID))
				fmt.Println(fmt.Sprintf("userEvent.ProfileURL: %s", userEvent.ProfileURL))
				fmt.Println(fmt.Sprintf("userEvent.Email: %s", userEvent.Email))
				fmt.Println(fmt.Sprintf("userEvent.Created: %s", userEvent.Created))
				fmt.Println(fmt.Sprintf("userEvent.Updated: %s", userEvent.Updated))
				fmt.Println(fmt.Sprintf("userEvent.ETag: %s", userEvent.ETag))
				fmt.Println(fmt.Sprintf("userEvent.Custom: %v", userEvent.Custom))

				if (userEvent.Event == pubnub.PNObjectsEventDelete) && (userEvent.UserID == userid) {
					mut.Lock()
					deleteUser = true
					mut.Unlock()
				}
				if (userEvent.Event == pubnub.PNObjectsEventUpdate) && (userEvent.UserID == userid) {
					mut.Lock()
					updateUser = true
					mut.Unlock()
				}
			case spaceEvent := <-listener.SpaceEvent:

				fmt.Println(" --- SpaceEvent: ")
				fmt.Println(fmt.Sprintf("%s", spaceEvent))
				fmt.Println(fmt.Sprintf("spaceEvent.Channel: %s", spaceEvent.Channel))
				fmt.Println(fmt.Sprintf("spaceEvent.SubscribedChannel: %s", spaceEvent.SubscribedChannel))
				fmt.Println(fmt.Sprintf("spaceEvent.Event: %s", spaceEvent.Event))
				fmt.Println(fmt.Sprintf("spaceEvent.SpaceID: %s", spaceEvent.SpaceID))
				fmt.Println(fmt.Sprintf("spaceEvent.Description: %s", spaceEvent.Description))
				fmt.Println(fmt.Sprintf("spaceEvent.Timestamp: %s", spaceEvent.Timestamp))
				fmt.Println(fmt.Sprintf("spaceEvent.Created: %s", spaceEvent.Created))
				fmt.Println(fmt.Sprintf("spaceEvent.Updated: %s", spaceEvent.Updated))
				fmt.Println(fmt.Sprintf("spaceEvent.ETag: %s", spaceEvent.ETag))
				fmt.Println(fmt.Sprintf("spaceEvent.Custom: %v", spaceEvent.Custom))
				if (spaceEvent.Event == pubnub.PNObjectsEventDelete) && (spaceEvent.SpaceID == spaceid) {
					mut.Lock()
					deleteSpace = true
					mut.Unlock()
				}
				if (spaceEvent.Event == pubnub.PNObjectsEventUpdate) && (spaceEvent.SpaceID == spaceid) {
					mut.Lock()
					updateSpace = true
					mut.Unlock()
				}

			case membershipEvent := <-listener.MembershipEvent:

				fmt.Println(" --- MembershipEvent: ")
				fmt.Println(fmt.Sprintf("%s", membershipEvent))
				fmt.Println(fmt.Sprintf("membershipEvent.Channel: %s", membershipEvent.Channel))
				fmt.Println(fmt.Sprintf("membershipEvent.SubscribedChannel: %s", membershipEvent.SubscribedChannel))
				fmt.Println(fmt.Sprintf("membershipEvent.Event: %s", membershipEvent.Event))
				fmt.Println(fmt.Sprintf("membershipEvent.SpaceID: %s", membershipEvent.SpaceID))
				fmt.Println(fmt.Sprintf("membershipEvent.UserID: %s", membershipEvent.UserID))
				fmt.Println(fmt.Sprintf("membershipEvent.Description: %s", membershipEvent.Description))
				fmt.Println(fmt.Sprintf("membershipEvent.Timestamp: %s", membershipEvent.Timestamp))
				fmt.Println(fmt.Sprintf("membershipEvent.Custom: %v", membershipEvent.Custom))
				if (membershipEvent.Event == pubnub.PNObjectsEventCreate) && (membershipEvent.SpaceID == spaceid) && (membershipEvent.UserID == userid) && ((membershipEvent.Channel == spaceid) || (membershipEvent.Channel == userid)) {
					mut.Lock()
					addUserToSpace = true
					mut.Unlock()
				}
				if (membershipEvent.Event == pubnub.PNObjectsEventCreate) && (membershipEvent.SpaceID == spaceid) && (membershipEvent.UserID == userid) && ((membershipEvent.Channel == spaceid) || (membershipEvent.Channel == userid)) {
					mut.Lock()
					addUserToSpace2 = true
					mut.Unlock()
				}
				if (membershipEvent.Event == pubnub.PNObjectsEventUpdate) && (membershipEvent.SpaceID == spaceid) && (membershipEvent.UserID == userid) && ((membershipEvent.Channel == spaceid) || (membershipEvent.Channel == userid)) {
					mut.Lock()
					updateUserMem = true
					mut.Unlock()
				}
				if (membershipEvent.Event == pubnub.PNObjectsEventUpdate) && (membershipEvent.SpaceID == spaceid) && (membershipEvent.UserID == userid) && ((membershipEvent.Channel == spaceid) || (membershipEvent.Channel == userid)) {
					mut.Lock()
					updateUserMem = true
					mut.Unlock()
				}
				if (membershipEvent.Event == pubnub.PNObjectsEventDelete) && (membershipEvent.SpaceID == spaceid) && (membershipEvent.UserID == userid) && ((membershipEvent.Channel == spaceid) || (membershipEvent.Channel == userid)) {
					mut.Lock()
					removeUserFromSpace = true
					mut.Unlock()
				}

			}
			fmt.Println("=>>>>>>>>>>>>> restart")
		}

	}()

	pnSub.AddListener(listener)

	pnSub.Subscribe().Channels([]string{userid, spaceid}).Execute()
	tic := time.NewTicker(time.Duration(eventWaitTime) * time.Second)
	select {
	case <-doneConnected:
	case <-tic.C:
		tic.Stop()
		assert.Fail("timeout")
	}

	name := "name"
	extid := "extid"
	purl := "profileurl"
	email := "email"
	desc := "desc"

	customUser := make(map[string]interface{})
	customUser["au"] = "bu"
	customUser["cu"] = "du"

	incl := []pubnub.PNUserSpaceInclude{
		pubnub.PNUserSpaceCustom,
	}

	//Create User
	_, st, err := pn.CreateUser().Include(incl).ID(userid).Name(name).ExternalID(extid).ProfileURL(purl).Email(email).Custom(customUser).Execute()
	assert.Nil(err)
	assert.Equal(200, st.StatusCode)

	//Create Space
	customSpace := make(map[string]interface{})
	customSpace["as"] = "bs"
	customSpace["cs"] = "ds"

	_, st4, err4 := pn.CreateSpace().Include(incl).ID(spaceid).Name(name).Description(desc).Custom(customSpace).Execute()
	assert.Nil(err4)
	assert.Equal(200, st4.StatusCode)

	time.Sleep(1 * time.Second)

	//Update User
	email = "email2"

	_, st2, err2 := pn.UpdateUser().Include(incl).ID(userid).Name(name).ExternalID(extid).ProfileURL(purl).Email(email).Custom(customUser).Execute()
	assert.Nil(err2)
	assert.Equal(200, st2.StatusCode)

	time.Sleep(1 * time.Second)
	mut.Lock()
	assert.True(updateUser)
	mut.Unlock()

	desc = "desc2"

	//Update Space
	_, st3, err3 := pn.UpdateSpace().Include(incl).ID(spaceid).Name(name).Description(desc).Custom(customSpace).Execute()
	assert.Nil(err3)
	assert.Equal(200, st3.StatusCode)

	time.Sleep(1 * time.Second)
	mut.Lock()
	assert.True(updateSpace)
	mut.Unlock()

	//Add user to space
	inclSm := []pubnub.PNMembersInclude{
		pubnub.PNMembersCustom,
		pubnub.PNMembersUser,
		pubnub.PNMembersUserCustom,
	}

	fmt.Println("inclSm===>", inclSm)
	for k, value := range inclSm {
		fmt.Println("inclSm===>", k, value)
	}

	custom3 := make(map[string]interface{})
	custom3["a3"] = "b3"
	custom3["c3"] = "d3"

	in := pubnub.PNMembersInput{
		ID:     userid,
		Custom: custom3,
	}

	inArr := []pubnub.PNMembersInput{
		in,
	}

	_, stAdd, errAdd := pn.ManageMembers().SpaceID(spaceid).Add(inArr).Update([]pubnub.PNMembersInput{}).Remove([]pubnub.PNMembersRemove{}).Include(inclSm).Limit(limit).Count(count).Execute()
	assert.Nil(errAdd)
	if errAdd != nil {
		fmt.Println("ManageMembers-->", errAdd)
	}
	assert.Equal(200, stAdd.StatusCode)

	time.Sleep(1 * time.Second)
	mut.Lock()
	assert.True(addUserToSpace && addUserToSpace2)
	mut.Unlock()

	//Update user membership

	//Read event

	custom5 := make(map[string]interface{})
	custom5["a5"] = "b5"
	custom5["c5"] = "d5"

	upMem := pubnub.PNMembershipsInput{
		ID:     spaceid,
		Custom: custom5,
	}

	upArrMem := []pubnub.PNMembershipsInput{
		upMem,
	}

	inclMemberships := []pubnub.PNMembershipsInclude{
		pubnub.PNMembershipsCustom,
		pubnub.PNMembershipsSpace,
		pubnub.PNMembershipsSpaceCustom,
	}

	resManageMemUp, stManageMemUp, errManageMemUp := pn.ManageMemberships().UserID(userid).Add([]pubnub.PNMembershipsInput{}).Update(upArrMem).Remove([]pubnub.PNMembershipsRemove{}).Include(inclMemberships).Limit(limit).Count(count).Execute()
	fmt.Println("resManageMemUp -->", resManageMemUp)
	assert.Nil(errManageMemUp)
	if errManageMemUp != nil {
		fmt.Println("ManageMemberships-->", errManageMemUp)
	}

	assert.Equal(200, stManageMemUp.StatusCode)

	time.Sleep(1 * time.Second)
	mut.Lock()
	assert.True(updateUserMem)
	mut.Unlock()

	//Remove user from space
	reMem := pubnub.PNMembershipsRemove{
		ID: spaceid,
	}

	reArrMem := []pubnub.PNMembershipsRemove{
		reMem,
	}
	_, stManageMemRem, errManageMemRem := pn.ManageMemberships().UserID(userid).Add([]pubnub.PNMembershipsInput{}).Update([]pubnub.PNMembershipsInput{}).Remove(reArrMem).Include(inclMemberships).Limit(limit).Count(count).Execute()
	assert.Nil(errManageMemRem)
	if errManageMemRem != nil {
		fmt.Println("ManageMemberships-->", errManageMemRem)
	}

	assert.Equal(200, stManageMemRem.StatusCode)

	time.Sleep(1 * time.Second)
	mut.Lock()
	assert.True(removeUserFromSpace)
	mut.Unlock()

	//Delete user
	res52, st52, err52 := pn.DeleteUser().ID(userid).Execute()
	assert.Nil(err52)
	assert.Equal(200, st52.StatusCode)
	assert.Nil(res52.Data)

	time.Sleep(1 * time.Second)
	mut.Lock()
	assert.True(deleteUser)
	mut.Unlock()

	//Delete Space
	res62, st62, err62 := pn.DeleteSpace().ID(spaceid).Execute()
	assert.Nil(err62)
	assert.Equal(200, st62.StatusCode)
	assert.Nil(res62.Data)

	time.Sleep(1 * time.Second)
	mut.Lock()
	assert.True(deleteSpace)
	mut.Unlock()
}
