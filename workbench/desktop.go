package main

// Desktop is the only object bound to the new client. Agent execution, local
// admin login and the legacy Mock Skill runner are deliberately not exposed.
// Credentials live only in the UI session; the server encrypts SSO credentials.
type Desktop struct{}

func (d *Desktop) AppName() string { return "智工 · 企业工作台" }
