package strmutil

func Seq(start, end, step int) ProcFunc {
	return func(p Proc) error {
		if start > end {
			for i := start; i >= end; i += step {
				if err := p.Send(i); err != nil {
					return err
				}
			}
			return nil
		}
		for i := start; i < end; i += step {
			if err := p.Send(i); err != nil {
				return err
			}
		}
		return nil
	}
}
