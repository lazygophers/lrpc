package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/atexit"
	"github.com/lazygophers/utils/candy"

	mysqlC "github.com/go-sql-driver/mysql"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
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
			// 检查字段是否存在于 FieldsByDBName 映射中
			field, exists := stmt.Schema.FieldsByDBName[dbName]
			if !exists {
				log.Errorf("field %s not found in FieldsByDBName for table %s", dbName, tabler.TableName())
				return fmt.Errorf("field %s not found in schema", dbName)
			}

			err = migrator.MigrateColumn(table, field, columnType)
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

	// 缓存 ParseIndexes() 结果，避免多次调用
	schemaIndexes := stmt.Schema.ParseIndexes()

	if len(indexList) == 0 && len(schemaIndexes) == 0 {
		// 表和 Schema 都没有索引，跳过索引对齐
		return nil
	}

	indexMap := make(map[string]gorm.Index, len(indexList))
	for _, index := range indexList {
		indexMap[index.Name()] = index
	}

	for _, dbIndex := range schemaIndexes {
		if index, ok := indexMap[dbIndex.Name]; ok {
			// 对齐一下字段是否相同
			if candy.SliceEqual(index.Columns(), candy.Map(dbIndex.Fields, func(t schema.IndexOption) string {
				return t.DBName
			})) {
				continue
			}

			// 检查是否为主键或唯一索引，避免重建导致数据问题
			if dbIndex.Class == "PRIMARY" || dbIndex.Type != "" {
				log.Warnf("skipping rebuild of special index %s (class: %s, type: %s) on table %s",
					dbIndex.Name, dbIndex.Class, dbIndex.Type, tabler.TableName())
				continue
			}

			// 通过事务创建
			log.Infof("rebuilding index %s on table %s due to column changes", dbIndex.Name, tabler.TableName())
			tx := session.Begin()

			// 检查事务是否成功开始
			if tx.Error != nil {
				log.Errorf("failed to begin transaction: %v", tx.Error)
				return tx.Error
			}

			err = tx.Migrator().DropIndex(table, index.Name())
			if err != nil {
				if rbErr := tx.Rollback().Error; rbErr != nil {
					log.Errorf("rollback failed: %v", rbErr)
				}
				log.Errorf("failed to drop index: %v", err)
				return err
			}

			err = tx.Migrator().CreateIndex(table, dbIndex.Name)
			if err != nil {
				if rbErr := tx.Rollback().Error; rbErr != nil {
					log.Errorf("rollback failed: %v", rbErr)
				}
				log.Errorf("failed to create index: %v", err)
				return err
			}

			err = tx.Commit().Error
			if err != nil {
				// Commit 失败时尝试回滚
				if rbErr := tx.Rollback().Error; rbErr != nil {
					log.Errorf("rollback after commit failure: %v", rbErr)
				}
				log.Errorf("failed to commit transaction: %v", err)
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
	if p.db == nil {
		log.Errorf("database connection is nil")
		return nil
	}
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
	if p.db == nil {
		log.Errorf("database connection is nil")
		return nil
	}
	return NewScoop(p.db, p.clientType)
}

func (p *Client) Ping() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *Client) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = sqlDB.Close()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}
