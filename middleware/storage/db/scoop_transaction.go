package db

// ——————————事务——————————

func (p *Scoop) Begin() *Scoop {
	return NewScoop(p._db.Begin(), p.clientType)
}

func (p *Scoop) Rollback() *Scoop {
	p._db.Rollback()
	return p
}

func (p *Scoop) Commit() *Scoop {
	p._db.Commit()
	return p
}

func (p *Scoop) CommitOrRollback(tx *Scoop, logic func(tx *Scoop) error) error {
	err := logic(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}
