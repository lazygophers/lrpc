package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/routine"

	//_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
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
	p := &Client{
		clientType: c.Type,
	}

	c.apply()

	if c.Logger == nil {
		c.Logger = GetDefaultLogger()
	}

	var d gorm.Dialector
	switch c.Type {
	case Sqlite:
		d = newSqlite(c)

	case MySQL:
		log.Infof("mysql://%s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
		d = mysql.New(mysql.Config{
			DSN: fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.Username, c.Password, c.Address, c.Port, c.Name),
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
			SkipInitializeWithVersion:     false,
			DefaultStringSize:             500,
			DefaultDatetimePrecision:      nil,
			DisableWithReturning:          false,
			DisableDatetimePrecision:      false,
			DontSupportRenameIndex:        false,
			DontSupportRenameColumn:       false,
			DontSupportForShareClause:     false,
			DontSupportNullAsDefaultValue: false,
			DontSupportRenameColumnUnique: false,
		})

		if c.Debug {
			_ = mysqlC.SetLogger(&mysqlLogger{})
		}

	//case "postgres":
	//	log.Infof("postgres://%s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
	//	d = postgres.New(postgres.Config{
	//		DSN:                  fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Address, c.Port, c.Username, c.Password, c.Name),
	//		PreferSimpleProtocol: true,
	//		WithoutReturning:     false,
	//		Conn:                 nil,
	//	})
	//
	//case "sqlserver":
	//	log.Infof("sqlserver://%s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
	//	d = sqlserver.Open(fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", c.Username, c.Password, c.Address, c.Port, c.Name))

	default:
		return nil, errors.New("unknown database")
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

		DisableNestedTransaction: false,

		AllowGlobalUpdate: true,
		CreateBatchSize:   100,
		TranslateError:    true,

		PropagateUnscoped: true,

		ClauseBuilders: nil,
		ConnPool:       nil,
		Dialector:      nil,
		Plugins:        nil,

		Logger: c.GormLogger,
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if c.Debug {
		p.db.Logger = c.Logger
	}

	if c.Debug {
		p.db = p.db.Debug()
	}

	//conn, err := p.db.DB()
	//if err != nil {
	//	log.Errorf("err:%v", err)
	//	return nil, err
	//}
	//
	//err = conn.Ping()
	//if err != nil {
	//	log.Errorf("err:%v", err)
	//	return nil, err
	//}

	err = p.AutoMigrates(tables...)
	if err != nil {
		log.Errorf("err:%v", err)
		return p, err
	}

	routine.GoWithMustSuccess(func() (err error) {
		switch c.Type {
		case "sqlite":
			// 自动减少存储文件大小
			err = p.db.Session(&gorm.Session{
				Initialized: true,
			}).Exec("PRAGMA auto_vacuum = 1").Error
			if err != nil {
				log.Errorf("err:%v", err)
			}
		}

		//err = p.AutoMigrates(tables...)
		//if err != nil {
		//	log.Errorf("err:%v", err)
		//	return err
		//}

		return nil
	})

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
	if x, ok := table.(Tabler); ok {
		log.Infof("auto migrate %s", x.TableName())
	}

	//ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	//defer cancel()

	session := p.db.
		Session(&gorm.Session{
			//NewDB: true,
			Initialized: true,
			//Context: ctx,
		})

	migrator := session.Migrator()

	tabler, ok := table.(Tabler)
	if !ok {
		log.Panicf("table is not Tabler")
		return errors.New("table is not Tabler")
	}

	if !migrator.HasTable(tabler.TableName()) {
		// 找不到，就创建表
		err = migrator.CreateTable(tabler)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	// 找到了，
	stmt := &gorm.Statement{
		DB:    session,
		Table: tabler.TableName(),
		Model: table,
	}
	if session.Statement != nil {
		stmt.TableExpr = session.Statement.TableExpr
	}

	err = stmt.ParseWithSpecialTableName(table, stmt.Table)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
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
			// TODO: 找到了，对比一下类型
			//log.Info(stmt.Schema.FieldsByDBName[dbName].GORMDataType)
			//log.Info(stmt.Schema.FieldsByDBName[dbName].DataType)
			//log.Info(columnType.ColumnType())

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

	indexMap := make(map[string]gorm.Index, len(indexList))
	for _, index := range indexList {
		indexMap[index.Name()] = index
	}

	for _, dbIndex := range stmt.Schema.ParseIndexes() {
		if index, ok := indexMap[dbIndex.Name]; ok {
			var needChange bool
			// 对齐一下字段是否相同
			if !candy.SliceEqual(index.Columns(), candy.Map(dbIndex.Fields, func(t schema.IndexOption) string {
				return t.DBName
			})) {
				needChange = true
			}

			if !needChange {
				continue
			}

			//// 判断一下，如果不是唯一索引且不是主键，就删除索引再重建，否则就输出 warn 日志
			//if value, ok := index.PrimaryKey(); value && ok {
			//	log.Warnf("skip change index %s of table %s, because it is a primary key", dbIndex.Name, tabler.TableName())
			//	continue
			//}
			//
			//if value, ok := index.Unique(); value && ok {
			//	log.Warnf("skip change unique index %s of table %s", dbIndex.Name, tabler.TableName())
			//	continue
			//}

			// 通过事务创建
			tx := session.Begin()
			err = tx.Migrator().DropIndex(table, index.Name())
			if err != nil {
				tx.Rollback()
				log.Errorf("err:%v", err)
				return err
			}

			err = tx.Migrator().CreateIndex(table, dbIndex.Name)
			if err != nil {
				tx.Rollback()
				log.Errorf("err:%v", err)
				return err
			}

			err = tx.Commit().Error
			if err != nil {
				tx.Rollback()
				log.Errorf("err:%v", err)
				return err
			}

		} else {
			log.Infof("try add index %s to %s", tabler.TableName(), dbIndex.Name)
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
		//NewDB:       true,
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
