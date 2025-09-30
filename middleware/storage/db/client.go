package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/candy"

	mysqlC "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Client struct {
	db *gorm.DB

	clientType string
}

func New(c *Config, tables ...interface{}) (*Client, error) {
	c.apply()

	p := &Client{
		clientType: c.Type,
	}

	if c.Logger == nil {
		c.Logger = GetDefaultLogger()
	}

	var d gorm.Dialector
	switch c.Type {
	case Sqlite:
		log.Infof("connecting to sqlite: %s", c.DSN())
		d = sqlite.Open(c.DSN())

	case MySQL:
		log.Infof("connecting to mysql: %s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
		d = mysql.New(mysql.Config{
			DSN: c.DSN(),
			DSNConfig: &mysqlC.Config{
				Timeout:                 time.Second * 5,
				ReadTimeout:             time.Second * 30,
				WriteTimeout:            time.Second * 30,
				AllowAllFiles:           true,
				AllowCleartextPasswords: true,
				AllowNativePasswords:    true,
				AllowOldPasswords:       true,
				CheckConnLiveness:       true,
				ClientFoundRows:         true,
				ColumnsWithAlias:        true,
				InterpolateParams:       true,
				MultiStatements:         true,
				ParseTime:               true,
			},
			DefaultStringSize: 500,
		})

		if c.Debug {
			_ = mysqlC.SetLogger(&mysqlLogger{})
		}

	default:
		return nil, fmt.Errorf("unknown database type: %s", c.Type)
	}

	var err error
	p.db, err = gorm.Open(d, &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: &schema.NamingStrategy{
			SingularTable: true,
		},
		FullSaveAssociations: true,

		PrepareStmt: true,

		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,

		AllowGlobalUpdate: true,
		CreateBatchSize:   100,
		TranslateError:    true,

		PropagateUnscoped: true,

		Logger: c.Logger,
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if c.Debug {
		p.db = p.db.Debug()
	}

	err = p.AutoMigrates(tables...)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// SQLite 特定优化：启用自动减少存储文件大小
	if c.Type == Sqlite {
		err = p.db.Session(&gorm.Session{
			Initialized: true,
		}).Exec("PRAGMA auto_vacuum = 1").Error
		if err != nil {
			log.Errorf("failed to set PRAGMA auto_vacuum: %v", err)
			// 不返回错误，因为这是优化设置，失败不应阻止初始化
		}
	}

	return p, nil
}

func (p *Client) AutoMigrates(dst ...interface{}) (err error) {
	for _, table := range dst {
		err = p.AutoMigrate(table)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	return nil
}

func (p *Client) AutoMigrate(table interface{}) (err error) {
	tabler, ok := table.(Tabler)
	if !ok {
		return fmt.Errorf("table type %T does not implement Tabler interface", table)
	}

	log.Infof("auto migrate %s", tabler.TableName())

	session := p.db.Session(&gorm.Session{
		Initialized: true,
	})

	migrator := session.Migrator()

	if !migrator.HasTable(tabler.TableName()) {
		// 找不到，就创建表
		err = migrator.CreateTable(tabler)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		// 新表创建成功，无需对齐字段和索引
		return nil
	}

	// 找到了，对齐字段和索引
	stmt := &gorm.Statement{
		DB:    session,
		Table: tabler.TableName(),
		Model: table,
	}
	// 复用现有 session 的 TableExpr（如果存在）
	if session.Statement != nil {
		stmt.TableExpr = session.Statement.TableExpr
	}

	err = stmt.ParseWithSpecialTableName(table, stmt.Table)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	if stmt.Schema == nil {
		log.Errorf("stmt.Schema is nil for table %s", tabler.TableName())
		return errors.New("stmt.Schema is nil")
	}

	// 对齐一下字段
	columnTypeList, err := migrator.ColumnTypes(tabler)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	columnTypeMap := make(map[string]gorm.ColumnType, len(columnTypeList))
	for _, columnType := range columnTypeList {
		columnTypeMap[columnType.Name()] = columnType
	}

	for _, dbName := range stmt.Schema.DBNames {
		if columnType, ok := columnTypeMap[dbName]; ok {
			err = migrator.MigrateColumn(table, stmt.Schema.FieldsByDBName[dbName], columnType)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		} else {
			// 找不到，所以要新建字段
			log.Infof("try add column %s to %s", dbName, tabler.TableName())
			err = migrator.AddColumn(table, dbName)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}
	}

	// 对齐一下索引
	indexList, err := migrator.GetIndexes(table)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	if len(indexList) == 0 && len(stmt.Schema.ParseIndexes()) == 0 {
		// 表和 Schema 都没有索引，跳过索引对齐
		return nil
	}

	indexMap := make(map[string]gorm.Index, len(indexList))
	for _, index := range indexList {
		indexMap[index.Name()] = index
	}

	for _, dbIndex := range stmt.Schema.ParseIndexes() {
		if index, ok := indexMap[dbIndex.Name]; ok {
			// 对齐一下字段是否相同
			if candy.SliceEqual(index.Columns(), candy.Map(dbIndex.Fields, func(t schema.IndexOption) string {
				return t.DBName
			})) {
				continue
			}

			// 通过事务创建
			log.Infof("rebuilding index %s on table %s due to column changes", dbIndex.Name, tabler.TableName())
			tx := session.Begin()
			err = tx.Migrator().DropIndex(table, index.Name())
			if err != nil {
				if rbErr := tx.Rollback().Error; rbErr != nil {
					log.Errorf("rollback failed: %v", rbErr)
				}
				log.Errorf("err:%v", err)
				return err
			}

			err = tx.Migrator().CreateIndex(table, dbIndex.Name)
			if err != nil {
				if rbErr := tx.Rollback().Error; rbErr != nil {
					log.Errorf("rollback failed: %v", rbErr)
				}
				log.Errorf("err:%v", err)
				return err
			}

			err = tx.Commit().Error
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

		} else {
			log.Infof("try add index %s to %s", dbIndex.Name, tabler.TableName())
			err = migrator.CreateIndex(table, dbIndex.Name)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}
		}
	}

	return nil
}

func (p *Client) Database() *gorm.DB {
	return p.db.Session(&gorm.Session{
		Initialized: true,
	})
}

func (p *Client) SqlDB() (*sql.DB, error) {
	db, err := p.db.DB()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return db, nil
}

func (p *Client) DriverType() string {
	return p.clientType
}

func (p *Client) NewScoop() *Scoop {
	return NewScoop(p.db, p.clientType)
}
