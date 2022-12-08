package apisvc

/*
func (s *Service) Cud(ctx context.Context, req *apipb.CUDRequest) (*apipb.CUDResponse, error) {
	p, err := s.p.GetPrice(ctx)
	if err != nil {
		return nil, err
	}

	// parse fee
	fee, ok := parseBigInt(req.Fee)
	if !ok {
		return nil, fmt.Errorf("invalid fee")
	}

	// check price is enough
	if fee.Cmp(p) < 0 {
		return nil, fmt.Errorf("price is not enough")
	}

	// generate id
	id := cudID(req)

	if id != req.Id {
		return nil, fmt.Errorf("invalid id")
	}

	// check signature
	valid, err := crypto.CheckSignature(req.From, req.Signature, []byte(id))
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("invalid signature")
	}

	// spend funds and then write data!

	//TODO: use DB name to map to connection string
	conStr, err := s.mp.GetConnectionInfo(req.From)
	if err != nil {
		return nil, err
	}

	db, err := sqlclient.Open(conStr)
	if err != nil {
		return nil, err
	}

	// req.Query is in the form table:CRUD, so parse those out
	qs := strings.Split(req.Query, ":")
	if len(qs) != 2 {
		return nil, fmt.Errorf("invalid query")
	}

	columns := make(map[string]any)
	for _, i := range req.Inputs {
		columns[i.Name] = i.Value
	}

	where := make(map[string]any)
	if req.Where.Name != "" {
		where[req.Where.Name] = req.Where.Value
	}

	switch qs[1] {
	case "insert":
		stmt := sqldriver.InsertStatement{
			Database: req.Database,
			Table:    qs[0],
			Input:    columns,
		}
		err = db.ExecuteInsert(ctx, stmt)
		if err != nil {
			return nil, err
		}
	case "update":
		stmt := sqldriver.UpdateStatement{
			Database: req.Database,
			Table:    qs[0],
			Input:    columns,
			Where:    where,
		}
		err = db.ExecuteUpdate(ctx, stmt)
		if err != nil {
			return nil, err
		}
	case "delete":
		stmt := sqldriver.DeleteStatement{
			Database: req.Database,
			Table:    qs[0],
			Where:    where,
		}
		err = db.ExecuteDelete(ctx, stmt)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid query")
	}

	return &apipb.CUDResponse{
		TraceId: "",
	}, nil
}

func (s *Service) Read(ctx context.Context, req *apipb.ReadRequest) (*apipb.ReadResponse, error) {
	where := make(map[string]any)
	if req.Where.Name != "" {
		where[req.Where.Name] = req.Where.Value
	}

	// parse out table name
	qs := strings.Split(req.Query, ":")
	if len(qs) != 2 {
		return nil, fmt.Errorf("invalid query")
	}

	stmt := sqldriver.SelectStatement{
		Database: req.Database,
		Table:    qs[0],
		Where:    where,
	}

	conStr, err := s.mp.GetConnectionInfo(req.Owner)
	if err != nil {
		return nil, err
	}

	db, err := sqlclient.Open(conStr)
	if err != nil {
		return nil, err
	}

	res, err := db.ExecuteSelect(ctx, stmt)
	if err != nil {
		return nil, err
	}

	ws := convertResult(res)

	return &apipb.ReadResponse{
		Result: ws,
	}, nil
}

func convertMaps(res []map[string]string) *apipb.ReadResponse {
	// iterate ov
	var rows []query.Row
	for _, c := range res {
		// iterate over columns
		var cols []query.Column
		for k, v := range c {
			cols = append(cols, query.Column{
				Name:  k,
				Value: sql.NullString{String: v},
			})
		}
		rows = append(rows, query.Row{
			Columns: cols,
		})
	}

	result := &query.Result{
		Rows: rows,
	}

	pbr := convertResult(result)

	return &apipb.ReadResponse{
		Result: pbr,
	}
}

func convertResultColumn(c *query.Column) *apipb.ColumnResult {
	return &apipb.ColumnResult{
		Name:  c.Name,
		Value: c.Value.String,
		Type:  convertType(c.Type),
	}
}

func convertResultRow(r *query.Row) *apipb.Row {
	var cols []*apipb.ColumnResult
	for _, c := range r.Columns {
		cols = append(cols, convertResultColumn(&c))
	}
	return &apipb.Row{
		Columns: cols,
	}
}

func convertResult(r *query.Result) *apipb.QueryResult {
	var rows []*apipb.Row
	for _, row := range r.Rows {
		rows = append(rows, convertResultRow(&row))
	}
	return &apipb.QueryResult{
		Rows: rows,
	}
}

func parseBigInt(s string) (*big.Int, bool) {
	return new(big.Int).SetString(s, 10)
}
*/
