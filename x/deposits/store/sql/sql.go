package sql

import (
    "database/sql"
    "fmt"
    "kwil/x/deposits/types"
    "math/big"
    "os"
    "strconv"
    "strings"

    _ "github.com/lib/pq"
)

type sqlstore struct {
    db *sql.DB
}

type SQLStore interface {
    Close() error
    Ping() error
    Exec(query string, args ...interface{}) (sql.Result, error)
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    SetHeight(h int64) error
    GetHeight() (int64, error)
    GetBalance(addr string) (*big.Int, error)
    GetSpent(addr string) (*big.Int, error)
    GetBalanceAndSpent(addr string) (string, string, error)
    CommitHeight(h int64) error
    CommitDeposits(h int64) error
    Expire(h int64) error
    Spend(addr string, amount string) error
    Deposit(txid, addr, amount string, h int64) error
    GetAllWithdrawals(h int64) ([]*types.WithdrawalRequest, error)
    StartWithdrawal(nonce, wallet, amount string, expiry int64) (*types.PendingWithdrawal, error)
    FinishWithdrawal(nonce string) (bool, error)
    RemoveBalance(addr string, amount string) error
    AddTx(string, string) error
    GetWithdrawalsForWallet(wallet string) ([]*types.PendingWithdrawal, error)
}

type SQLConfig struct {
    Host     string
    Port     int64
    User     string
    Password string
    Database string
    SSL      bool
}

func New(sc *SQLConfig) (*sqlstore, error) {
    ssl := "disable"
    if sc.SSL {
        ssl = "require"
    }

    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=%s", sc.Host, sc.Port, sc.User, sc.Password, sc.Database, ssl)

    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        return nil, err
    }

    return &sqlstore{
        db: db,
    }, nil
}

func TestDB() (*sqlstore, error) {

    db, err := New(&SQLConfig{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "password",
        Database: "postgres",
        SSL:      false,
    })
    if err != nil {
        return nil, err
    }

    // execute initialization script
    // read in test_init.sql

    c, err := os.ReadFile("./test_init.sql")
    if err != nil {
        return nil, err
    }
    initSql := string(c)

    _, err = db.Exec(initSql)
    if err != nil {
        return nil, err
    }

    return db, nil
}

func (s *sqlstore) Close() error {
    return s.db.Close()
}

func (s *sqlstore) Ping() error {
    return s.db.Ping()
}

func (s *sqlstore) Exec(query string, args ...interface{}) (sql.Result, error) {
    return s.db.Exec(query, args...)
}

func (s *sqlstore) Query(query string, args ...interface{}) (*sql.Rows, error) {
    return s.db.Query(query, args...)
}

func (s *sqlstore) QueryRow(query string, args ...interface{}) *sql.Row {
    return s.db.QueryRow(query, args...)
}

func (s *sqlstore) SetHeight(h int64) error {
    _, err := s.Exec(" SELECT set_height($1);", h)
    return err
}

func (s *sqlstore) GetHeight() (int64, error) {
    var h int64
    err := s.QueryRow("SELECT get_height()").Scan(&h)
    return h, err
}

func (s *sqlstore) GetBalance(addr string) (*big.Int, error) {
    var bStr string
    err := s.QueryRow("SELECT get_balance($1)", addr).Scan(&bStr)
    if err != nil {
        return big.NewInt(0), nil
    }

    return parseBigInt(bStr)
}

func (s *sqlstore) GetSpent(addr string) (*big.Int, error) {
    var bStr string
    err := s.QueryRow("SELECT get_spent($1)", addr).Scan(&bStr)
    if err != nil {
        return big.NewInt(0), nil
    }

    return parseBigInt(bStr)
}

func (s *sqlstore) CommitHeight(h int64) error {
    // start transaction
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    _, err = tx.Exec("SELECT commit_height($1)", h)
    if err != nil {
        return tx.Rollback()
    }
    return tx.Commit()
}

func (s *sqlstore) Spend(addr string, amount string) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    _, err = tx.Exec("SELECT spend_money($1, $2)", addr, amount)
    if err != nil {
        err = tx.Rollback()
        if err != nil { // rollback likely won't fail, but just in case
            return err
        }
        return ErrInsufficientFunds
    }
    return tx.Commit()
}

func (s *sqlstore) Deposit(txid, addr string, amount string, h int64) error {
    _, err := s.Exec("SELECT deposit($1, $2, $3, $4)", txid, addr, amount, h)
    return err
}

func (s *sqlstore) GetBalanceAndSpent(addr string) (string, string, error) {
    var b, sp string
    err := s.QueryRow("SELECT get_balance($1), get_spent($1)", addr).Scan(&b, &sp)
    return b, sp, err
}

func (s *sqlstore) GetAllWithdrawals(h int64) ([]*types.WithdrawalRequest, error) {
    var ret []*types.WithdrawalRequest
    res, err := s.Query("SELECT get_all_withdrawals($1)", h)
    if err != nil {
        return nil, err
    }

    /*
    	res should have columns:
    	- nonce, wallet, amount, fee, expiry
    */

    for res.Next() {
        // should change nonce to cid, but am unsure if this would break res.Scan
        var nonce, wallet, amount, fee string
        var expiry int64
        err := res.Scan(&nonce, &wallet, &amount, &fee, &expiry)
        if err != nil {
            return nil, err
        }

        // append to ret
        ret = append(ret, &types.WithdrawalRequest{
            Cid:        nonce,
            Wallet:     wallet,
            Amount:     amount,
            Spent:      fee,
            Expiration: expiry,
        })
    }

    return ret, err
}

func (s *sqlstore) RemoveBalance(addr string, amount string) error {
    _, err := s.Exec("SELECT remove_balance($1, $2)", addr, amount)
    return err
}

// will begin the withdrawal process.  This will create a withdrawal request and return the nonce
// TODO: this should return info regarding the spent amount (fee)
func (s *sqlstore) StartWithdrawal(nonce, wallet, amount string, expiry int64) (*types.PendingWithdrawal, error) {
    tx, err := s.db.Begin()
    if err != nil {
        return nil, err
    }
    res, err := tx.Query("SELECT start_withdrawal($1, $2, $3, $4);", wallet, nonce, amount, expiry)
    if err != nil {
        return nil, tx.Rollback()
    }

    res.Next()
    var resp string

    // returned as address, nonce, amount, fee, expiry (1,2,3,4,5)
    err = res.Scan(&resp)
    if err != nil {
        return nil, err
    }

    err = tx.Commit()
    if err != nil {
        return nil, err
    }

    // parse resp
    // trim off the first and last characters
    resp = resp[1 : len(resp)-1]

    // split on commas
    split := strings.Split(resp, ",")

    // parse the final value to int64
    exp, err := strconv.ParseInt(split[4], 10, 64)
    if err != nil {
        return nil, err
    }

    var famt, ffee string
    if split[2] == "NULL" {
        famt = "0"
    } else {
        famt = split[2]
    }

    if split[3] == "NULL" {
        ffee = "0"
    } else {
        ffee = split[3]
    }

    return &types.PendingWithdrawal{
        Wallet:     split[0],
        Cid:        split[1],
        Amount:     famt,
        Fee:        ffee,
        Expiration: exp,
    }, nil

}

// will delete the withdrawal by the nonce.  This is called when we have heard back from the blockchain
func (s *sqlstore) FinishWithdrawal(nonce string) (bool, error) {
    tx, err := s.db.Begin()
    if err != nil {
        return false, err
    }
    res, err := tx.Query("SELECT finish_withdrawal($1)", nonce)
    if err != nil {
        return false, tx.Rollback()
    }

    res.Next()
    var resp bool
    err = res.Scan(&resp)
    if err != nil {
        return false, err
    }

    err = tx.Commit()
    if err != nil {
        return false, err
    }

    return resp, nil
}

func (s *sqlstore) CommitDeposits(h int64) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }

    _, err = tx.Exec("SELECT commit_deposits($1);", h)
    if err != nil {
        return tx.Rollback()
    }

    return tx.Commit()
}

func (s *sqlstore) Expire(h int64) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    _, err = tx.Exec("SELECT expire($1)", h)
    if err != nil {
        return tx.Rollback()
    }

    return tx.Commit()
}

// should only be used in tests
func (s *sqlstore) RemoveWallet(addr string) error {
    fmt.Println("WARNING: this should only be used in tests, and never in production")
    rows, err := s.Query("SELECT wallet_id FROM wallets WHERE wallet = '" + addr + "';")
    if err != nil {
        return err
    }
    var wallet_id int64

    // should only be one row
    for rows.Next() {
        err := rows.Scan(&wallet_id)
        if err != nil {
            return err
        }
    }

    _, err = s.Exec("DELETE FROM wallets WHERE wallet_id = '" + fmt.Sprint(wallet_id) + "';")
    if err != nil {
        return err
    }
    _, err = s.Exec("DELETE FROM deposits WHERE wallet = '" + addr + "';")
    if err != nil {
        return err
    }
    _, err = s.Exec("DELETE FROM withdrawals WHERE wallet_id = '" + fmt.Sprint(wallet_id) + "';")

    return err
}

func (s *sqlstore) AddTx(cid string, tx string) error {
    _, err := s.Exec("SELECT add_tx($1, $2)", cid, tx)
    return err
}

func (s *sqlstore) GetWithdrawalsForWallet(w string) ([]*types.PendingWithdrawal, error) {
    res, err := s.db.Query("SELECT get_withdrawals_addr($1)", w)
    if err != nil {
        return nil, err
    }
    fmt.Println(res)

    var wds []*types.PendingWithdrawal
    // now we loop through the results and append them to the slice
    for res.Next() {
        var wd types.PendingWithdrawal
        wd.Wallet = w
        var ur string
        err := res.Scan(&ur)
        if err != nil {
            return nil, err
        }

        // trim off the first and last characters
        ur = ur[1 : len(ur)-1]

        // split on commas
        split := strings.Split(ur, ",")
        wd.Cid = split[0]
        wd.Amount = split[1]
        wd.Fee = split[2]
        wd.Expiration, err = strconv.ParseInt(split[3], 10, 64)
        if err != nil {
            return nil, err
        }
        wd.Tx = split[4]

        wds = append(wds, &wd)
    }

    return wds, nil
}

func parseBigInt(amt string) (*big.Int, error) {
    b := big.NewInt(0)
    _, ok := b.SetString(amt, 10)
    if !ok {
        return nil, ErrFailedToParseBigInt
    }

    return b, nil
}
