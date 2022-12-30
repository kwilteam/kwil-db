package executables

func (d *databaseInterface) CanExecute(wallet, query string) bool {

	// check if the default roles have permission
	for _, role := range d.DefaultRoles {
		if d.Access[role] == query {
			return true
		}
	}

	// check if wallet has permission
	return d.Owner == wallet

	// since we do not currently have ways of defining non-default roles, I will not implement any more logic here
}
