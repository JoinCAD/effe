package effe

import (
	"github.com/cockroachdb/apd"
)

func sum(rp rangeProvider, args []value) (value, error) {
	var summand = apd.New(0, 0)
	for _, v := range args {
		switch rv := v.(type) {
		case rangeValue:
			err := rp.iterate(rv.r, func(cv value) error {
				switch t := cv.(type) {
				case numberValue:
					// TODO: don't discard condition
					_, err := apd.BaseContext.Add(summand, summand, t.v)
					return err
				default:
					return nil
				}
				return nil
			})

			if err != nil {
				return nullValue{}, err
			}
		}
	}

	return numberValue{
		v: summand,
	}, nil
}
