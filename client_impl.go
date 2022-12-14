package gohbase

import (
	"context"

	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/hrpc"
	"go.uber.org/atomic"
)

var _ Client = &client{}

type client struct {
	*session

	cli   gohbase.Client
	clone *atomic.Bool
}

func (c *client) Context(ctx context.Context) Client {
	tx := c.getInstance()
	tx.WithContext(ctx)

	return tx
}

func (c *client) Table(table string) Client {
	tx := c.getInstance()
	tx.WithTable(table)

	return tx
}

func (c *client) Key(key string) Client {
	tx := c.getInstance()
	tx.WithKey(key)

	return tx
}

func (c *client) Family(family string) Client {
	tx := c.getInstance()
	tx.WithFamily(family)

	return tx
}

func (c *client) Qualifier(qualifier string) Client {
	tx := c.getInstance()
	tx.WithQualifier(qualifier)

	return tx
}

func (c *client) Amount(amount int64) Client {
	tx := c.getInstance()
	tx.WithAmount(amount)

	return tx
}

func (c *client) Range(startRow, stopRow string) Client {
	tx := c.getInstance()
	tx.WithRange(startRow, stopRow)

	return tx
}

func (c *client) Values(values map[string]map[string][]byte) Client {
	tx := c.getInstance()
	tx.WithValues(values)

	return tx
}

func (c *client) ExpectedValue(expectedValue []byte) Client {
	tx := c.getInstance()
	tx.WithExpectedValue(expectedValue)

	return tx
}

func (c *client) Options(opts ...func(hrpc.Call) error) Client {
	tx := c.getInstance()
	tx.WithOptions(opts...)

	return tx
}

func (c *client) Scan() Result {
	tx := c.getInstance()
	if err := tx.validate(); err != nil {
		return &result{
			err: err,
		}
	}

	var (
		scan *hrpc.Scan
		err  error
	)

	if tx.isSetRange() {
		scan, err = hrpc.NewScanRangeStr(tx.ctx, tx.table, *tx.startRow, *tx.stopRow, tx.opts...)
	} else {
		scan, err = hrpc.NewScanStr(tx.ctx, tx.table, tx.opts...)
	}

	if err != nil {
		return &result{
			err: err,
		}
	}

	return &result{
		scanner: tx.cli.Scan(scan),
	}
}

func (c *client) Get() Result {
	tx := c.getInstance()
	if err := tx.session.validate(); err != nil {
		return &result{
			err: err,
		}
	}

	get, err := hrpc.NewGetStr(tx.ctx, tx.table, tx.key, tx.opts...)
	if err != nil {
		return &result{
			err: err,
		}
	}

	r := &result{}
	r.result, r.err = tx.cli.Get(get)

	return r
}

func (c *client) Put() Result {
	tx := c.getInstance()
	if err := tx.session.validate(); err != nil {
		return &result{
			err: err,
		}
	}

	put, err := hrpc.NewPutStr(tx.ctx, tx.table, tx.key, tx.values, tx.opts...)
	if err != nil {
		return &result{
			err: err,
		}
	}

	r := &result{}
	r.result, r.err = tx.cli.Put(put)

	return r
}

func (c *client) Delete() Result {
	tx := c.getInstance()
	if err := tx.session.validate(); err != nil {
		return &result{
			err: err,
		}
	}

	del, err := hrpc.NewDelStr(tx.ctx, tx.table, tx.key, tx.values, tx.opts...)
	if err != nil {
		return &result{
			err: err,
		}
	}

	r := &result{}
	r.result, r.err = tx.cli.Delete(del)

	return r
}

func (c *client) Append() Result {
	tx := c.getInstance()
	if err := tx.session.validate(); err != nil {
		return &result{
			err: err,
		}
	}

	app, err := hrpc.NewAppStr(tx.ctx, tx.table, tx.key, tx.values, tx.opts...)
	if err != nil {
		return &result{
			err: err,
		}
	}

	r := &result{}
	r.result, r.err = tx.cli.Append(app)

	return r
}

func (c *client) Increment() Result {
	tx := c.getInstance()
	if err := tx.validate(); err != nil {
		return &result{
			err: err,
		}
	}

	var (
		inc *hrpc.Mutate
		err error
	)

	if tx.isSetAmount() {
		inc, err = hrpc.NewIncStrSingle(tx.ctx, tx.table, tx.key, *tx.family, *tx.qualifier, *tx.amount, tx.opts...)
	} else {
		inc, err = hrpc.NewIncStr(tx.ctx, tx.table, tx.key, tx.values, tx.opts...)
	}
	if err != nil {
		return &result{
			err: err,
		}
	}

	r := &result{}
	r.i64, r.err = tx.cli.Increment(inc)

	return r
}

func (c *client) CheckAndPut() Result {
	tx := c.getInstance()
	if err := tx.session.validate(); err != nil {
		return &result{
			err: err,
		}
	}

	put, err := hrpc.NewPutStr(tx.ctx, tx.table, tx.key, tx.values, tx.opts...)
	if err != nil {
		return &result{
			err: err,
		}
	}

	r := &result{}
	r.b, r.err = tx.cli.CheckAndPut(put, *tx.family, *tx.qualifier, tx.expectedValue)

	return r
}

func (c *client) getInstance() *client {
	if c.clone.Load() {
		return c
	}

	return &client{
		cli:     c.cli,
		session: newSession(context.Background()),
		clone:   atomic.NewBool(true),
	}
}

func (c *client) Close() {
	c.cli.Close()
}

func NewClient(opts ...Option) Client {
	o := newOption()
	o.apply(opts...)
	hClient := gohbase.NewClient(o.addr, o.gohbaseOpts...)

	return &client{
		cli:   hClient,
		clone: atomic.NewBool(false),
	}
}
