package effe

import (
	"github.com/cockroachdb/apd"
)

func sum(rp rangeProvider, args []value) (value, error) {
	var summand = apd.New(0, 0)
	for _, v := range args {
		err := rp.iterate(v.rangeValue, func(cv value) error {
			if cv.numberValue != nil {
				// TODO: don't discard condition
				_, err := apd.BaseContext.Add(summand, summand, cv.numberValue)
				return err
			}
			return nil
		})
		if err != nil {
			return value{}, err
		}
	}

	return value{
		numberValue: summand,
	}, nil
}
