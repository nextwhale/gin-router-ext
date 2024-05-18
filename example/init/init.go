package init

import (
	gre "github.com/nextwhale/gin-router-ext"
	ga "github.com/nextwhale/go-acl"
)


// engine
var eng *gre.EngineExt

func App() *gre.EngineExt {
	if eng == nil {
		eng = gre.New(nil)
	}
	return eng
}

// acl module
var acl *ga.ACL

func ACL() *ga.ACL {
	if acl == nil {
		LoadACL()
	}
	return acl
}

func LoadACL(){
	roleAdmin := ga.NewRoleWithUniquePermissions("1", "编辑组", []string{"/admin/admin/list", "/admin/admin/edit/:id", "/admin/admin/del/:id"})
	roleEditor := ga.NewRoleWithUniquePermissions("editor", "编辑组", []string{"/admin/article/list", "/admin/article/edit/:id", "/admin/article/del/:id"})
	roleVideoAuditor := ga.NewRoleWithUniquePermissions("video_auditor", "视频审核员", []string{"/admin/video/list", "/admin/video/edit/:id", "/admin/video/del/:id"})

	// add roles
	_acl := &ga.ACL{}
	_acl.AddRole(roleAdmin, roleEditor, roleVideoAuditor)
	acl = _acl
}

