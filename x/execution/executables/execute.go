package executables

func (d *executableInterface) CanExecute(wallet, query string) bool {

	// check if the default roles have permission
	for _, role := range d.DefaultRoles {
		_, ok := d.Access[role][query]
		if ok {
			return true
		}
	}

	// check if wallet has permission
	return d.Owner == wallet

	// since we do not currently have ways of defining non-default roles, I will not implement any more logic here
}
