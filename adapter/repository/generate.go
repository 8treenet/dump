package repository

import (
	"fmt"
	"github.com/8treenet/dump/domain/po"
	"github.com/8treenet/freedom"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

// GORMRepository .
type GORMRepository interface {
	db() *gorm.DB
	GetWorker() freedom.Worker
}

// Builder .
type Builder interface {
	Execute(db *gorm.DB, object interface{}) error
}

// Pager .
type Pager struct {
	pageSize  int
	page      int
	totalPage int
	fields    []string
	orders    []string
}

// NewDescPager .
func NewDescPager(column string, columns ...string) *Pager {
	return newDefaultPager("desc", column, columns...)
}

// NewAscPager .
func NewAscPager(column string, columns ...string) *Pager {
	return newDefaultPager("asc", column, columns...)
}

// NewDescOrder .
func newDefaultPager(sort, field string, args ...string) *Pager {
	fields := []string{field}
	fields = append(fields, args...)
	orders := []string{}
	for index := 0; index < len(fields); index++ {
		orders = append(orders, sort)
	}
	return &Pager{
		fields: fields,
		orders: orders,
	}
}

// Order .
func (p *Pager) Order() interface{} {
	if len(p.fields) == 0 {
		return nil
	}
	args := []string{}
	for index := 0; index < len(p.fields); index++ {
		args = append(args, fmt.Sprintf("`%s` %s", p.fields[index], p.orders[index]))
	}

	return strings.Join(args, ",")
}

// TotalPage .
func (p *Pager) TotalPage() int {
	return p.totalPage
}

// SetPage .
func (p *Pager) SetPage(page, pageSize int) *Pager {
	p.page = page
	p.pageSize = pageSize
	return p
}

// Execute .
func (p *Pager) Execute(db *gorm.DB, object interface{}) (e error) {
	pageFind := false
	orderValue := p.Order()
	if orderValue != nil {
		db = db.Order(orderValue)
	} else {
		db = db.Set("gorm:order_by_primary_key", "DESC")
	}
	if p.page != 0 && p.pageSize != 0 {
		pageFind = true
		db = db.Offset((p.page - 1) * p.pageSize).Limit(p.pageSize)
	}

	resultDB := db.Find(object)
	if resultDB.Error != nil {
		return resultDB.Error
	}

	if !pageFind {
		return
	}

	var count int
	e = resultDB.Offset(0).Limit(1).Count(&count).Error
	if e == nil && count != 0 {
		//计算分页
		if count%p.pageSize == 0 {
			p.totalPage = count / p.pageSize
		} else {
			p.totalPage = count/p.pageSize + 1
		}
	}
	return
}

func ormErrorLog(repo GORMRepository, model, method string, e error, expression ...interface{}) {
	if e == nil || e == gorm.ErrRecordNotFound {
		return
	}
	repo.GetWorker().Logger().Errorf("Orm error, model: %s, method: %s, expression :%v, reason for error:%v", model, method, expression, e)
}

// findOrderLog .
func findOrderLog(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderLog", "findOrderLog", e, now)
		ormErrorLog(repo, "OrderLog", "findOrderLog", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderLogListByPrimarys .
func findOrderLogListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("OrderLog", "findOrderLogListByPrimarys", e, now)
	ormErrorLog(repo, "OrderLog", "findOrderLogsByPrimarys", e, primarys)
	return
}

// findOrderLogByWhere .
func findOrderLogByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderLog", "findOrderLogByWhere", e, now)
		ormErrorLog(repo, "OrderLog", "findOrderLogByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderLogByMap .
func findOrderLogByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderLog", "findOrderLogByMap", e, now)
		ormErrorLog(repo, "OrderLog", "findOrderLogByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderLogList .
func findOrderLogList(repo GORMRepository, query po.OrderLog, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderLog", "findOrderLogList", e, now)
		ormErrorLog(repo, "OrderLog", "findOrderLogs", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findOrderLogListByWhere .
func findOrderLogListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderLog", "findOrderLogListByWhere", e, now)
		ormErrorLog(repo, "OrderLog", "findOrderLogsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findOrderLogListByMap .
func findOrderLogListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderLog", "findOrderLogListByMap", e, now)
		ormErrorLog(repo, "OrderLog", "findOrderLogsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createOrderLog .
func createOrderLog(repo GORMRepository, object *po.OrderLog) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("OrderLog", "createOrderLog", e, now)
	ormErrorLog(repo, "OrderLog", "createOrderLog", e, *object)
	return
}

// saveOrderLog .
func saveOrderLog(repo GORMRepository, object *po.OrderLog) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("OrderLog", "saveOrderLog", e, now)
	ormErrorLog(repo, "OrderLog", "saveOrderLog", e, *object)
	return
}

// findAlbums .
func findAlbums(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Albums", "findAlbums", e, now)
		ormErrorLog(repo, "Albums", "findAlbums", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findAlbumsListByPrimarys .
func findAlbumsListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Albums", "findAlbumsListByPrimarys", e, now)
	ormErrorLog(repo, "Albums", "findAlbumssByPrimarys", e, primarys)
	return
}

// findAlbumsByWhere .
func findAlbumsByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Albums", "findAlbumsByWhere", e, now)
		ormErrorLog(repo, "Albums", "findAlbumsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findAlbumsByMap .
func findAlbumsByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Albums", "findAlbumsByMap", e, now)
		ormErrorLog(repo, "Albums", "findAlbumsByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findAlbumsList .
func findAlbumsList(repo GORMRepository, query po.Albums, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Albums", "findAlbumsList", e, now)
		ormErrorLog(repo, "Albums", "findAlbumss", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findAlbumsListByWhere .
func findAlbumsListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Albums", "findAlbumsListByWhere", e, now)
		ormErrorLog(repo, "Albums", "findAlbumssByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findAlbumsListByMap .
func findAlbumsListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Albums", "findAlbumsListByMap", e, now)
		ormErrorLog(repo, "Albums", "findAlbumssByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createAlbums .
func createAlbums(repo GORMRepository, object *po.Albums) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Albums", "createAlbums", e, now)
	ormErrorLog(repo, "Albums", "createAlbums", e, *object)
	return
}

// saveAlbums .
func saveAlbums(repo GORMRepository, object *po.Albums) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Albums", "saveAlbums", e, now)
	ormErrorLog(repo, "Albums", "saveAlbums", e, *object)
	return
}

// findCart .
func findCart(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCart", e, now)
		ormErrorLog(repo, "Cart", "findCart", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findCartListByPrimarys .
func findCartListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Cart", "findCartListByPrimarys", e, now)
	ormErrorLog(repo, "Cart", "findCartsByPrimarys", e, primarys)
	return
}

// findCartByWhere .
func findCartByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartByWhere", e, now)
		ormErrorLog(repo, "Cart", "findCartByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findCartByMap .
func findCartByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartByMap", e, now)
		ormErrorLog(repo, "Cart", "findCartByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findCartList .
func findCartList(repo GORMRepository, query po.Cart, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartList", e, now)
		ormErrorLog(repo, "Cart", "findCarts", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findCartListByWhere .
func findCartListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartListByWhere", e, now)
		ormErrorLog(repo, "Cart", "findCartsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findCartListByMap .
func findCartListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Cart", "findCartListByMap", e, now)
		ormErrorLog(repo, "Cart", "findCartsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createCart .
func createCart(repo GORMRepository, object *po.Cart) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Cart", "createCart", e, now)
	ormErrorLog(repo, "Cart", "createCart", e, *object)
	return
}

// saveCart .
func saveCart(repo GORMRepository, object *po.Cart) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Cart", "saveCart", e, now)
	ormErrorLog(repo, "Cart", "saveCart", e, *object)
	return
}

// findDelivery .
func findDelivery(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDelivery", e, now)
		ormErrorLog(repo, "Delivery", "findDelivery", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findDeliveryListByPrimarys .
func findDeliveryListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryListByPrimarys", e, now)
	ormErrorLog(repo, "Delivery", "findDeliverysByPrimarys", e, primarys)
	return
}

// findDeliveryByWhere .
func findDeliveryByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryByWhere", e, now)
		ormErrorLog(repo, "Delivery", "findDeliveryByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findDeliveryByMap .
func findDeliveryByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryByMap", e, now)
		ormErrorLog(repo, "Delivery", "findDeliveryByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findDeliveryList .
func findDeliveryList(repo GORMRepository, query po.Delivery, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryList", e, now)
		ormErrorLog(repo, "Delivery", "findDeliverys", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findDeliveryListByWhere .
func findDeliveryListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryListByWhere", e, now)
		ormErrorLog(repo, "Delivery", "findDeliverysByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findDeliveryListByMap .
func findDeliveryListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Delivery", "findDeliveryListByMap", e, now)
		ormErrorLog(repo, "Delivery", "findDeliverysByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createDelivery .
func createDelivery(repo GORMRepository, object *po.Delivery) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Delivery", "createDelivery", e, now)
	ormErrorLog(repo, "Delivery", "createDelivery", e, *object)
	return
}

// saveDelivery .
func saveDelivery(repo GORMRepository, object *po.Delivery) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Delivery", "saveDelivery", e, now)
	ormErrorLog(repo, "Delivery", "saveDelivery", e, *object)
	return
}

// findDump .
func findDump(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Dump", "findDump", e, now)
		ormErrorLog(repo, "Dump", "findDump", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findDumpListByPrimarys .
func findDumpListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Dump", "findDumpListByPrimarys", e, now)
	ormErrorLog(repo, "Dump", "findDumpsByPrimarys", e, primarys)
	return
}

// findDumpByWhere .
func findDumpByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Dump", "findDumpByWhere", e, now)
		ormErrorLog(repo, "Dump", "findDumpByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findDumpByMap .
func findDumpByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Dump", "findDumpByMap", e, now)
		ormErrorLog(repo, "Dump", "findDumpByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findDumpList .
func findDumpList(repo GORMRepository, query po.Dump, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Dump", "findDumpList", e, now)
		ormErrorLog(repo, "Dump", "findDumps", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findDumpListByWhere .
func findDumpListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Dump", "findDumpListByWhere", e, now)
		ormErrorLog(repo, "Dump", "findDumpsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findDumpListByMap .
func findDumpListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Dump", "findDumpListByMap", e, now)
		ormErrorLog(repo, "Dump", "findDumpsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createDump .
func createDump(repo GORMRepository, object *po.Dump) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Dump", "createDump", e, now)
	ormErrorLog(repo, "Dump", "createDump", e, *object)
	return
}

// saveDump .
func saveDump(repo GORMRepository, object *po.Dump) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Dump", "saveDump", e, now)
	ormErrorLog(repo, "Dump", "saveDump", e, *object)
	return
}

// findGoods .
func findGoods(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoods", e, now)
		ormErrorLog(repo, "Goods", "findGoods", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findGoodsListByPrimarys .
func findGoodsListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsListByPrimarys", e, now)
	ormErrorLog(repo, "Goods", "findGoodssByPrimarys", e, primarys)
	return
}

// findGoodsByWhere .
func findGoodsByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsByWhere", e, now)
		ormErrorLog(repo, "Goods", "findGoodsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findGoodsByMap .
func findGoodsByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsByMap", e, now)
		ormErrorLog(repo, "Goods", "findGoodsByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findGoodsList .
func findGoodsList(repo GORMRepository, query po.Goods, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsList", e, now)
		ormErrorLog(repo, "Goods", "findGoodss", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findGoodsListByWhere .
func findGoodsListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsListByWhere", e, now)
		ormErrorLog(repo, "Goods", "findGoodssByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findGoodsListByMap .
func findGoodsListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Goods", "findGoodsListByMap", e, now)
		ormErrorLog(repo, "Goods", "findGoodssByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createGoods .
func createGoods(repo GORMRepository, object *po.Goods) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Goods", "createGoods", e, now)
	ormErrorLog(repo, "Goods", "createGoods", e, *object)
	return
}

// saveGoods .
func saveGoods(repo GORMRepository, object *po.Goods) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Goods", "saveGoods", e, now)
	ormErrorLog(repo, "Goods", "saveGoods", e, *object)
	return
}

// findOrder .
func findOrder(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrder", e, now)
		ormErrorLog(repo, "Order", "findOrder", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderListByPrimarys .
func findOrderListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Order", "findOrderListByPrimarys", e, now)
	ormErrorLog(repo, "Order", "findOrdersByPrimarys", e, primarys)
	return
}

// findOrderByWhere .
func findOrderByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderByWhere", e, now)
		ormErrorLog(repo, "Order", "findOrderByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderByMap .
func findOrderByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderByMap", e, now)
		ormErrorLog(repo, "Order", "findOrderByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderList .
func findOrderList(repo GORMRepository, query po.Order, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderList", e, now)
		ormErrorLog(repo, "Order", "findOrders", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findOrderListByWhere .
func findOrderListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderListByWhere", e, now)
		ormErrorLog(repo, "Order", "findOrdersByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findOrderListByMap .
func findOrderListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Order", "findOrderListByMap", e, now)
		ormErrorLog(repo, "Order", "findOrdersByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createOrder .
func createOrder(repo GORMRepository, object *po.Order) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Order", "createOrder", e, now)
	ormErrorLog(repo, "Order", "createOrder", e, *object)
	return
}

// saveOrder .
func saveOrder(repo GORMRepository, object *po.Order) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Order", "saveOrder", e, now)
	ormErrorLog(repo, "Order", "saveOrder", e, *object)
	return
}

// findOrderDetail .
func findOrderDetail(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetail", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetail", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderDetailListByPrimarys .
func findOrderDetailListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailListByPrimarys", e, now)
	ormErrorLog(repo, "OrderDetail", "findOrderDetailsByPrimarys", e, primarys)
	return
}

// findOrderDetailByWhere .
func findOrderDetailByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailByWhere", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetailByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderDetailByMap .
func findOrderDetailByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailByMap", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetailByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findOrderDetailList .
func findOrderDetailList(repo GORMRepository, query po.OrderDetail, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailList", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetails", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findOrderDetailListByWhere .
func findOrderDetailListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailListByWhere", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetailsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findOrderDetailListByMap .
func findOrderDetailListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("OrderDetail", "findOrderDetailListByMap", e, now)
		ormErrorLog(repo, "OrderDetail", "findOrderDetailsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createOrderDetail .
func createOrderDetail(repo GORMRepository, object *po.OrderDetail) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("OrderDetail", "createOrderDetail", e, now)
	ormErrorLog(repo, "OrderDetail", "createOrderDetail", e, *object)
	return
}

// saveOrderDetail .
func saveOrderDetail(repo GORMRepository, object *po.OrderDetail) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("OrderDetail", "saveOrderDetail", e, now)
	ormErrorLog(repo, "OrderDetail", "saveOrderDetail", e, *object)
	return
}

// findProduct .
func findProduct(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Product", "findProduct", e, now)
		ormErrorLog(repo, "Product", "findProduct", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findProductListByPrimarys .
func findProductListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Product", "findProductListByPrimarys", e, now)
	ormErrorLog(repo, "Product", "findProductsByPrimarys", e, primarys)
	return
}

// findProductByWhere .
func findProductByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Product", "findProductByWhere", e, now)
		ormErrorLog(repo, "Product", "findProductByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findProductByMap .
func findProductByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Product", "findProductByMap", e, now)
		ormErrorLog(repo, "Product", "findProductByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findProductList .
func findProductList(repo GORMRepository, query po.Product, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Product", "findProductList", e, now)
		ormErrorLog(repo, "Product", "findProducts", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findProductListByWhere .
func findProductListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Product", "findProductListByWhere", e, now)
		ormErrorLog(repo, "Product", "findProductsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findProductListByMap .
func findProductListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Product", "findProductListByMap", e, now)
		ormErrorLog(repo, "Product", "findProductsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createProduct .
func createProduct(repo GORMRepository, object *po.Product) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Product", "createProduct", e, now)
	ormErrorLog(repo, "Product", "createProduct", e, *object)
	return
}

// saveProduct .
func saveProduct(repo GORMRepository, object *po.Product) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Product", "saveProduct", e, now)
	ormErrorLog(repo, "Product", "saveProduct", e, *object)
	return
}

// findTestUsers .
func findTestUsers(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestUsers", "findTestUsers", e, now)
		ormErrorLog(repo, "TestUsers", "findTestUsers", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findTestUsersListByPrimarys .
func findTestUsersListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("TestUsers", "findTestUsersListByPrimarys", e, now)
	ormErrorLog(repo, "TestUsers", "findTestUserssByPrimarys", e, primarys)
	return
}

// findTestUsersByWhere .
func findTestUsersByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestUsers", "findTestUsersByWhere", e, now)
		ormErrorLog(repo, "TestUsers", "findTestUsersByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findTestUsersByMap .
func findTestUsersByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestUsers", "findTestUsersByMap", e, now)
		ormErrorLog(repo, "TestUsers", "findTestUsersByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findTestUsersList .
func findTestUsersList(repo GORMRepository, query po.TestUsers, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestUsers", "findTestUsersList", e, now)
		ormErrorLog(repo, "TestUsers", "findTestUserss", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findTestUsersListByWhere .
func findTestUsersListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestUsers", "findTestUsersListByWhere", e, now)
		ormErrorLog(repo, "TestUsers", "findTestUserssByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findTestUsersListByMap .
func findTestUsersListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestUsers", "findTestUsersListByMap", e, now)
		ormErrorLog(repo, "TestUsers", "findTestUserssByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createTestUsers .
func createTestUsers(repo GORMRepository, object *po.TestUsers) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("TestUsers", "createTestUsers", e, now)
	ormErrorLog(repo, "TestUsers", "createTestUsers", e, *object)
	return
}

// saveTestUsers .
func saveTestUsers(repo GORMRepository, object *po.TestUsers) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("TestUsers", "saveTestUsers", e, now)
	ormErrorLog(repo, "TestUsers", "saveTestUsers", e, *object)
	return
}

// findAdmin .
func findAdmin(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdmin", e, now)
		ormErrorLog(repo, "Admin", "findAdmin", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findAdminListByPrimarys .
func findAdminListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminListByPrimarys", e, now)
	ormErrorLog(repo, "Admin", "findAdminsByPrimarys", e, primarys)
	return
}

// findAdminByWhere .
func findAdminByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminByWhere", e, now)
		ormErrorLog(repo, "Admin", "findAdminByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findAdminByMap .
func findAdminByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminByMap", e, now)
		ormErrorLog(repo, "Admin", "findAdminByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findAdminList .
func findAdminList(repo GORMRepository, query po.Admin, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminList", e, now)
		ormErrorLog(repo, "Admin", "findAdmins", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findAdminListByWhere .
func findAdminListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminListByWhere", e, now)
		ormErrorLog(repo, "Admin", "findAdminsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findAdminListByMap .
func findAdminListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("Admin", "findAdminListByMap", e, now)
		ormErrorLog(repo, "Admin", "findAdminsByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createAdmin .
func createAdmin(repo GORMRepository, object *po.Admin) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("Admin", "createAdmin", e, now)
	ormErrorLog(repo, "Admin", "createAdmin", e, *object)
	return
}

// saveAdmin .
func saveAdmin(repo GORMRepository, object *po.Admin) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("Admin", "saveAdmin", e, now)
	ormErrorLog(repo, "Admin", "saveAdmin", e, *object)
	return
}

// findTestEmails .
func findTestEmails(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestEmails", "findTestEmails", e, now)
		ormErrorLog(repo, "TestEmails", "findTestEmails", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findTestEmailsListByPrimarys .
func findTestEmailsListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("TestEmails", "findTestEmailsListByPrimarys", e, now)
	ormErrorLog(repo, "TestEmails", "findTestEmailssByPrimarys", e, primarys)
	return
}

// findTestEmailsByWhere .
func findTestEmailsByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestEmails", "findTestEmailsByWhere", e, now)
		ormErrorLog(repo, "TestEmails", "findTestEmailsByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findTestEmailsByMap .
func findTestEmailsByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestEmails", "findTestEmailsByMap", e, now)
		ormErrorLog(repo, "TestEmails", "findTestEmailsByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findTestEmailsList .
func findTestEmailsList(repo GORMRepository, query po.TestEmails, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestEmails", "findTestEmailsList", e, now)
		ormErrorLog(repo, "TestEmails", "findTestEmailss", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findTestEmailsListByWhere .
func findTestEmailsListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestEmails", "findTestEmailsListByWhere", e, now)
		ormErrorLog(repo, "TestEmails", "findTestEmailssByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findTestEmailsListByMap .
func findTestEmailsListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("TestEmails", "findTestEmailsListByMap", e, now)
		ormErrorLog(repo, "TestEmails", "findTestEmailssByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createTestEmails .
func createTestEmails(repo GORMRepository, object *po.TestEmails) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("TestEmails", "createTestEmails", e, now)
	ormErrorLog(repo, "TestEmails", "createTestEmails", e, *object)
	return
}

// saveTestEmails .
func saveTestEmails(repo GORMRepository, object *po.TestEmails) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("TestEmails", "saveTestEmails", e, now)
	ormErrorLog(repo, "TestEmails", "saveTestEmails", e, *object)
	return
}

// findUser .
func findUser(repo GORMRepository, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUser", e, now)
		ormErrorLog(repo, "User", "findUser", e, result)
	}()
	db := repo.db()
	if len(builders) == 0 {
		e = db.Where(result).Last(result).Error
		return
	}
	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findUserListByPrimarys .
func findUserListByPrimarys(repo GORMRepository, results interface{}, primarys ...interface{}) (e error) {
	now := time.Now()
	e = repo.db().Find(results, primarys).Error
	freedom.Prometheus().OrmWithLabelValues("User", "findUserListByPrimarys", e, now)
	ormErrorLog(repo, "User", "findUsersByPrimarys", e, primarys)
	return
}

// findUserByWhere .
func findUserByWhere(repo GORMRepository, query string, args []interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserByWhere", e, now)
		ormErrorLog(repo, "User", "findUserByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findUserByMap .
func findUserByMap(repo GORMRepository, query map[string]interface{}, result interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserByMap", e, now)
		ormErrorLog(repo, "User", "findUserByMap", e, query)
	}()

	db := repo.db().Where(query)
	if len(builders) == 0 {
		e = db.Last(result).Error
		return
	}

	e = builders[0].Execute(db.Limit(1), result)
	return
}

// findUserList .
func findUserList(repo GORMRepository, query po.User, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserList", e, now)
		ormErrorLog(repo, "User", "findUsers", e, query)
	}()
	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findUserListByWhere .
func findUserListByWhere(repo GORMRepository, query string, args []interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserListByWhere", e, now)
		ormErrorLog(repo, "User", "findUsersByWhere", e, query, args)
	}()
	db := repo.db()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// findUserListByMap .
func findUserListByMap(repo GORMRepository, query map[string]interface{}, results interface{}, builders ...Builder) (e error) {
	now := time.Now()
	defer func() {
		freedom.Prometheus().OrmWithLabelValues("User", "findUserListByMap", e, now)
		ormErrorLog(repo, "User", "findUsersByMap", e, query)
	}()

	db := repo.db().Where(query)

	if len(builders) == 0 {
		e = db.Find(results).Error
		return
	}
	e = builders[0].Execute(db, results)
	return
}

// createUser .
func createUser(repo GORMRepository, object *po.User) (rowsAffected int64, e error) {
	now := time.Now()
	db := repo.db().Create(object)
	rowsAffected = db.RowsAffected
	e = db.Error
	freedom.Prometheus().OrmWithLabelValues("User", "createUser", e, now)
	ormErrorLog(repo, "User", "createUser", e, *object)
	return
}

// saveUser .
func saveUser(repo GORMRepository, object *po.User) (affected int64, e error) {
	now := time.Now()
	db := repo.db().Model(object).Updates(object.TakeChanges())
	e = db.Error
	affected = db.RowsAffected
	freedom.Prometheus().OrmWithLabelValues("User", "saveUser", e, now)
	ormErrorLog(repo, "User", "saveUser", e, *object)
	return
}
