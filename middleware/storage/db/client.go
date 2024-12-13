package db

import (
	"errors"
	"fmt"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/routine"
	"reflect"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	mysqlC "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
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
	case "sqlite":
		d = newSqlite(c)

	case "mysql":
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

		_ = mysqlC.SetLogger(&mysqlLogger{})

	case "postgres":
		log.Infof("postgres://%s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
		d = postgres.New(postgres.Config{
			DSN:                  fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.Address, c.Port, c.Username, c.Password, c.Name),
			PreferSimpleProtocol: true,
			WithoutReturning:     false,
			Conn:                 nil,
		})

	case "sqlserver":
		log.Infof("sqlserver://%s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
		d = sqlserver.Open(fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", c.Username, c.Password, c.Address, c.Port, c.Name))

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
		Logger:               c.Logger,

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
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
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

		err = p.AutoMigrates(tables...)
		if err != nil {
			log.Errorf("err:%v", err)
		}

		return nil
	})

	return p, nil
}

func (p *Client) AutoMigrates(dst ...interface{}) (err error) {
	for _, table := range dst {
		err = p.AutoMigrate(table)
		if err != nil {
			log.Errorf("err:%v", err)
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
		}).Migrator()

	err = session.AutoMigrate(table)
	if err != nil {
		log.Errorf("err:%v", err)

		switch x := err.(type) {
		case *mysqlC.MySQLError:
			// do something here
		default:
			log.Errorf("err:%v", reflect.TypeOf(x))
		}

		if t, ok := table.(Tabler); ok {
			log.Errorf("table: %s", t.TableName())
		}

		return err
	}

	return nil
}

func (p *Client) Database() *gorm.DB {
	return p.db.Session(&gorm.Session{
		//NewDB:       true,
		Initialized: true,
	})
}

func (p *Client) NewScoop() *Scoop {
	return NewScoop(p.db)
}
