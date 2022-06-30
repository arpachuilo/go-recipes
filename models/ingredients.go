// Code generated by SQLBoiler 4.10.2 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Ingredient is an object representing the database table.
type Ingredient struct {
	ID         null.Int64  `boil:"id" json:"id,omitempty" toml:"id" yaml:"id,omitempty"`
	Recipeid   null.Int64  `boil:"recipeid" json:"recipeid,omitempty" toml:"recipeid" yaml:"recipeid,omitempty"`
	Ingredient null.String `boil:"ingredient" json:"ingredient,omitempty" toml:"ingredient" yaml:"ingredient,omitempty"`

	R *ingredientR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L ingredientL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var IngredientColumns = struct {
	ID         string
	Recipeid   string
	Ingredient string
}{
	ID:         "id",
	Recipeid:   "recipeid",
	Ingredient: "ingredient",
}

var IngredientTableColumns = struct {
	ID         string
	Recipeid   string
	Ingredient string
}{
	ID:         "ingredients.id",
	Recipeid:   "ingredients.recipeid",
	Ingredient: "ingredients.ingredient",
}

// Generated where

type whereHelpernull_Int64 struct{ field string }

func (w whereHelpernull_Int64) EQ(x null.Int64) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Int64) NEQ(x null.Int64) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Int64) LT(x null.Int64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Int64) LTE(x null.Int64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Int64) GT(x null.Int64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Int64) GTE(x null.Int64) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpernull_Int64) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Int64) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

type whereHelpernull_String struct{ field string }

func (w whereHelpernull_String) EQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_String) NEQ(x null.String) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_String) LT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_String) LTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_String) GT(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_String) GTE(x null.String) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpernull_String) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_String) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

var IngredientWhere = struct {
	ID         whereHelpernull_Int64
	Recipeid   whereHelpernull_Int64
	Ingredient whereHelpernull_String
}{
	ID:         whereHelpernull_Int64{field: "\"ingredients\".\"id\""},
	Recipeid:   whereHelpernull_Int64{field: "\"ingredients\".\"recipeid\""},
	Ingredient: whereHelpernull_String{field: "\"ingredients\".\"ingredient\""},
}

// IngredientRels is where relationship names are stored.
var IngredientRels = struct {
	RecipeidRecipe string
}{
	RecipeidRecipe: "RecipeidRecipe",
}

// ingredientR is where relationships are stored.
type ingredientR struct {
	RecipeidRecipe *Recipe `boil:"RecipeidRecipe" json:"RecipeidRecipe" toml:"RecipeidRecipe" yaml:"RecipeidRecipe"`
}

// NewStruct creates a new relationship struct
func (*ingredientR) NewStruct() *ingredientR {
	return &ingredientR{}
}

// ingredientL is where Load methods for each relationship are stored.
type ingredientL struct{}

var (
	ingredientAllColumns            = []string{"id", "recipeid", "ingredient"}
	ingredientColumnsWithoutDefault = []string{}
	ingredientColumnsWithDefault    = []string{"id", "recipeid", "ingredient"}
	ingredientPrimaryKeyColumns     = []string{"id"}
	ingredientGeneratedColumns      = []string{"id"}
)

type (
	// IngredientSlice is an alias for a slice of pointers to Ingredient.
	// This should almost always be used instead of []Ingredient.
	IngredientSlice []*Ingredient
	// IngredientHook is the signature for custom Ingredient hook methods
	IngredientHook func(context.Context, boil.ContextExecutor, *Ingredient) error

	ingredientQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	ingredientType                 = reflect.TypeOf(&Ingredient{})
	ingredientMapping              = queries.MakeStructMapping(ingredientType)
	ingredientPrimaryKeyMapping, _ = queries.BindMapping(ingredientType, ingredientMapping, ingredientPrimaryKeyColumns)
	ingredientInsertCacheMut       sync.RWMutex
	ingredientInsertCache          = make(map[string]insertCache)
	ingredientUpdateCacheMut       sync.RWMutex
	ingredientUpdateCache          = make(map[string]updateCache)
	ingredientUpsertCacheMut       sync.RWMutex
	ingredientUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var ingredientAfterSelectHooks []IngredientHook

var ingredientBeforeInsertHooks []IngredientHook
var ingredientAfterInsertHooks []IngredientHook

var ingredientBeforeUpdateHooks []IngredientHook
var ingredientAfterUpdateHooks []IngredientHook

var ingredientBeforeDeleteHooks []IngredientHook
var ingredientAfterDeleteHooks []IngredientHook

var ingredientBeforeUpsertHooks []IngredientHook
var ingredientAfterUpsertHooks []IngredientHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Ingredient) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range ingredientAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Ingredient) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range ingredientBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Ingredient) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range ingredientAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Ingredient) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range ingredientBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Ingredient) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range ingredientAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Ingredient) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range ingredientBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Ingredient) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range ingredientAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Ingredient) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range ingredientBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Ingredient) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range ingredientAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddIngredientHook registers your hook function for all future operations.
func AddIngredientHook(hookPoint boil.HookPoint, ingredientHook IngredientHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		ingredientAfterSelectHooks = append(ingredientAfterSelectHooks, ingredientHook)
	case boil.BeforeInsertHook:
		ingredientBeforeInsertHooks = append(ingredientBeforeInsertHooks, ingredientHook)
	case boil.AfterInsertHook:
		ingredientAfterInsertHooks = append(ingredientAfterInsertHooks, ingredientHook)
	case boil.BeforeUpdateHook:
		ingredientBeforeUpdateHooks = append(ingredientBeforeUpdateHooks, ingredientHook)
	case boil.AfterUpdateHook:
		ingredientAfterUpdateHooks = append(ingredientAfterUpdateHooks, ingredientHook)
	case boil.BeforeDeleteHook:
		ingredientBeforeDeleteHooks = append(ingredientBeforeDeleteHooks, ingredientHook)
	case boil.AfterDeleteHook:
		ingredientAfterDeleteHooks = append(ingredientAfterDeleteHooks, ingredientHook)
	case boil.BeforeUpsertHook:
		ingredientBeforeUpsertHooks = append(ingredientBeforeUpsertHooks, ingredientHook)
	case boil.AfterUpsertHook:
		ingredientAfterUpsertHooks = append(ingredientAfterUpsertHooks, ingredientHook)
	}
}

// One returns a single ingredient record from the query.
func (q ingredientQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Ingredient, error) {
	o := &Ingredient{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for ingredients")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Ingredient records from the query.
func (q ingredientQuery) All(ctx context.Context, exec boil.ContextExecutor) (IngredientSlice, error) {
	var o []*Ingredient

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Ingredient slice")
	}

	if len(ingredientAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Ingredient records in the query.
func (q ingredientQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count ingredients rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q ingredientQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if ingredients exists")
	}

	return count > 0, nil
}

// RecipeidRecipe pointed to by the foreign key.
func (o *Ingredient) RecipeidRecipe(mods ...qm.QueryMod) recipeQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.Recipeid),
	}

	queryMods = append(queryMods, mods...)

	return Recipes(queryMods...)
}

// LoadRecipeidRecipe allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (ingredientL) LoadRecipeidRecipe(ctx context.Context, e boil.ContextExecutor, singular bool, maybeIngredient interface{}, mods queries.Applicator) error {
	var slice []*Ingredient
	var object *Ingredient

	if singular {
		object = maybeIngredient.(*Ingredient)
	} else {
		slice = *maybeIngredient.(*[]*Ingredient)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &ingredientR{}
		}
		if !queries.IsNil(object.Recipeid) {
			args = append(args, object.Recipeid)
		}

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &ingredientR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.Recipeid) {
					continue Outer
				}
			}

			if !queries.IsNil(obj.Recipeid) {
				args = append(args, obj.Recipeid)
			}

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`recipes`),
		qm.WhereIn(`recipes.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Recipe")
	}

	var resultSlice []*Recipe
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Recipe")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for recipes")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for recipes")
	}

	if len(ingredientAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.RecipeidRecipe = foreign
		if foreign.R == nil {
			foreign.R = &recipeR{}
		}
		foreign.R.RecipeidIngredients = append(foreign.R.RecipeidIngredients, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.Recipeid, foreign.ID) {
				local.R.RecipeidRecipe = foreign
				if foreign.R == nil {
					foreign.R = &recipeR{}
				}
				foreign.R.RecipeidIngredients = append(foreign.R.RecipeidIngredients, local)
				break
			}
		}
	}

	return nil
}

// SetRecipeidRecipe of the ingredient to the related item.
// Sets o.R.RecipeidRecipe to related.
// Adds o to related.R.RecipeidIngredients.
func (o *Ingredient) SetRecipeidRecipe(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Recipe) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"ingredients\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 0, []string{"recipeid"}),
		strmangle.WhereClause("\"", "\"", 0, ingredientPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	queries.Assign(&o.Recipeid, related.ID)
	if o.R == nil {
		o.R = &ingredientR{
			RecipeidRecipe: related,
		}
	} else {
		o.R.RecipeidRecipe = related
	}

	if related.R == nil {
		related.R = &recipeR{
			RecipeidIngredients: IngredientSlice{o},
		}
	} else {
		related.R.RecipeidIngredients = append(related.R.RecipeidIngredients, o)
	}

	return nil
}

// RemoveRecipeidRecipe relationship.
// Sets o.R.RecipeidRecipe to nil.
// Removes o from all passed in related items' relationships struct.
func (o *Ingredient) RemoveRecipeidRecipe(ctx context.Context, exec boil.ContextExecutor, related *Recipe) error {
	var err error

	queries.SetScanner(&o.Recipeid, nil)
	if _, err = o.Update(ctx, exec, boil.Whitelist("recipeid")); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	if o.R != nil {
		o.R.RecipeidRecipe = nil
	}
	if related == nil || related.R == nil {
		return nil
	}

	for i, ri := range related.R.RecipeidIngredients {
		if queries.Equal(o.Recipeid, ri.Recipeid) {
			continue
		}

		ln := len(related.R.RecipeidIngredients)
		if ln > 1 && i < ln-1 {
			related.R.RecipeidIngredients[i] = related.R.RecipeidIngredients[ln-1]
		}
		related.R.RecipeidIngredients = related.R.RecipeidIngredients[:ln-1]
		break
	}
	return nil
}

// Ingredients retrieves all the records using an executor.
func Ingredients(mods ...qm.QueryMod) ingredientQuery {
	mods = append(mods, qm.From("\"ingredients\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"ingredients\".*"})
	}

	return ingredientQuery{q}
}

// FindIngredient retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindIngredient(ctx context.Context, exec boil.ContextExecutor, iD null.Int64, selectCols ...string) (*Ingredient, error) {
	ingredientObj := &Ingredient{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"ingredients\" where \"id\"=?", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, ingredientObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from ingredients")
	}

	if err = ingredientObj.doAfterSelectHooks(ctx, exec); err != nil {
		return ingredientObj, err
	}

	return ingredientObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Ingredient) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no ingredients provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(ingredientColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	ingredientInsertCacheMut.RLock()
	cache, cached := ingredientInsertCache[key]
	ingredientInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			ingredientAllColumns,
			ingredientColumnsWithDefault,
			ingredientColumnsWithoutDefault,
			nzDefaults,
		)
		wl = strmangle.SetComplement(wl, ingredientGeneratedColumns)

		cache.valueMapping, err = queries.BindMapping(ingredientType, ingredientMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(ingredientType, ingredientMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"ingredients\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"ingredients\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into ingredients")
	}

	if !cached {
		ingredientInsertCacheMut.Lock()
		ingredientInsertCache[key] = cache
		ingredientInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Ingredient.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Ingredient) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	ingredientUpdateCacheMut.RLock()
	cache, cached := ingredientUpdateCache[key]
	ingredientUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			ingredientAllColumns,
			ingredientPrimaryKeyColumns,
		)
		wl = strmangle.SetComplement(wl, ingredientGeneratedColumns)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update ingredients, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"ingredients\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 0, wl),
			strmangle.WhereClause("\"", "\"", 0, ingredientPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(ingredientType, ingredientMapping, append(wl, ingredientPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update ingredients row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for ingredients")
	}

	if !cached {
		ingredientUpdateCacheMut.Lock()
		ingredientUpdateCache[key] = cache
		ingredientUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q ingredientQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for ingredients")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for ingredients")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o IngredientSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), ingredientPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"ingredients\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, ingredientPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in ingredient slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all ingredient")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Ingredient) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no ingredients provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(ingredientColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	ingredientUpsertCacheMut.RLock()
	cache, cached := ingredientUpsertCache[key]
	ingredientUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			ingredientAllColumns,
			ingredientColumnsWithDefault,
			ingredientColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			ingredientAllColumns,
			ingredientPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("models: unable to upsert ingredients, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(ingredientPrimaryKeyColumns))
			copy(conflict, ingredientPrimaryKeyColumns)
		}
		cache.query = buildUpsertQuerySQLite(dialect, "\"ingredients\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(ingredientType, ingredientMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(ingredientType, ingredientMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "models: unable to upsert ingredients")
	}

	if !cached {
		ingredientUpsertCacheMut.Lock()
		ingredientUpsertCache[key] = cache
		ingredientUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Ingredient record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Ingredient) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Ingredient provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), ingredientPrimaryKeyMapping)
	sql := "DELETE FROM \"ingredients\" WHERE \"id\"=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from ingredients")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for ingredients")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q ingredientQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no ingredientQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from ingredients")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for ingredients")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o IngredientSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(ingredientBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), ingredientPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"ingredients\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, ingredientPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from ingredient slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for ingredients")
	}

	if len(ingredientAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Ingredient) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindIngredient(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *IngredientSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := IngredientSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), ingredientPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"ingredients\".* FROM \"ingredients\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, ingredientPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in IngredientSlice")
	}

	*o = slice

	return nil
}

// IngredientExists checks if the Ingredient row exists.
func IngredientExists(ctx context.Context, exec boil.ContextExecutor, iD null.Int64) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"ingredients\" where \"id\"=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if ingredients exists")
	}

	return exists, nil
}
