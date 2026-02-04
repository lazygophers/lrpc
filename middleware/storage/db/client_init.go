package db

import (
	"fmt"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	gormLog "gorm.io/gorm/logger"

	mysqlC "github.com/go-sql-driver/mysql"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func New(c *Config, tables ...interface{}) (*Client, error) {
	c.apply()

	// Check if mock mode is enabled
	if c.Mock {
		client, err := newMock(c, tables...)
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}
		return client, nil
	}

	// 确保 JSON 序列化器已注册
	ensureSerializerRegistered()

	p := &Client{
		clientType: c.Type,
	}

	if c.Logger == nil {
		c.Logger = GetDefaultLogger()
	}

	var d gorm.Dialector
	switch c.Type {
	case Sqlite:
		if c.Password != "" {
			// Password is set, try to use CGO version for encryption
			if hasCGOSupport {
				log.Infof("connecting to sqlite with encryption (CGO/SQLCipher): %s", c.DSN())
			} else {
				log.Warnf("SQLite password is set but CGO is not enabled. Encryption is NOT active. To enable encryption, rebuild with CGO_ENABLED=1")
				log.Infof("connecting to sqlite (no CGO, no encryption): %s", c.DSN())
			}
		} else {
			// No password, use pure Go version
			log.Infof("connecting to sqlite (no CGO): %s", c.DSN())
		}

		d = newSqliteDialector(c.DSN())

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
			err := mysqlC.SetLogger(&mysqlLogger{})
			if err != nil {
				log.Errorf("failed to set mysql logger: %v", err)
			}
		}

	case Postgres:
		log.Infof("connecting to postgres: %s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
		d = postgres.New(postgres.Config{
			DSN:                  c.DSN(),
			PreferSimpleProtocol: true,
		})

	case ClickHouse:
		log.Infof("connecting to clickhouse: %s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
		d = clickhouse.Open(c.DSN())

	case TiDB:
		log.Infof("connecting to tidb: %s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
		// TiDB is MySQL-compatible, use MySQL driver
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
			err := mysqlC.SetLogger(&mysqlLogger{})
			if err != nil {
				log.Errorf("failed to set tidb logger: %v", err)
			}
		}

	case GaussDB:
		log.Infof("connecting to gaussdb: %s:******@%s:%d/%s", c.Username, c.Address, c.Port, c.Name)
		// GaussDB is PostgreSQL-compatible, use PostgreSQL driver
		d = postgres.New(postgres.Config{
			DSN:                  c.DSN(),
			PreferSimpleProtocol: true,
		})

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

		TranslateError: true,

		PropagateUnscoped: true,

		// Disable GORM's built-in logger to avoid duplicate SQL logging
		// Our custom logger (GetDefaultLogger()) is used in Scoop methods instead
		Logger: gormLog.Discard,
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	atexit.Register(func() {
		err := p.Close()
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	})

	if c.Debug {
		p.db = p.db.Debug()
	}

	sqlDb, err := p.SqlDB()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if c.MaxOpenConns > 0 {
		sqlDb.SetMaxOpenConns(c.MaxOpenConns)
	} else if c.MaxOpenConns < 0 {
		sqlDb.SetMaxOpenConns(0)
	}

	if c.MaxIdleConns > 0 {
		sqlDb.SetMaxIdleConns(c.MaxIdleConns)
	} else if c.MaxIdleConns < 0 {
		sqlDb.SetMaxIdleConns(0)
	}

	err = p.AutoMigrates(tables...)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// SQLite 特定优化：启用自动减少存储文件大小
	// 注意：PRAGMA auto_vacuum 仅在新数据库创建时生效
	// 对已存在的数据库，需要执行 VACUUM 命令才能应用此设置
	if c.Type == Sqlite {
		err = p.db.Session(&gorm.Session{
			Initialized: true,
		}).Exec("PRAGMA auto_vacuum = INCREMENTAL").Error
		if err != nil {
			log.Errorf("failed to set PRAGMA auto_vacuum: %v", err)
			// 不返回错误，因为这是优化设置，失败不应阻止初始化
		}
	}

	// 注册退出时自动关闭数据库连接
	atexit.Register(func() {
		err := p.Close()
		if err != nil {
			log.Errorf("err:%v", err)
			return
		}
	})

	return p, nil
}
