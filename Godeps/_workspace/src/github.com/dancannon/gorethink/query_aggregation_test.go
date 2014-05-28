package gorethink

import (
	test "launchpad.net/gocheck"
)

func (s *RethinkSuite) TestAggregationReduce(c *test.C) {
	var response int
	query := Expr(arr).Reduce(func(acc, val RqlTerm) RqlTerm {
		return acc.Add(val)
	})
	r, err := query.RunRow(sess)
	c.Assert(err, test.IsNil)

	err = r.Scan(&response)
	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 45)
}

func (s *RethinkSuite) TestAggregationExprCount(c *test.C) {
	var response int
	query := Expr(arr).Count()
	r, err := query.RunRow(sess)
	c.Assert(err, test.IsNil)

	err = r.Scan(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, 9)
}

func (s *RethinkSuite) TestAggregationDistinct(c *test.C) {
	var response []int
	query := Expr(darr).Distinct()
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.HasLen, 5)
}

func (s *RethinkSuite) TestAggregationGroupMapReduce(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group(func(row RqlTerm) RqlTerm {
		return row.Field("id").Mod(2).Eq(0)
	}).Map(func(row RqlTerm) RqlTerm {
		return row.Field("num")
	}).Reduce(func(acc, num RqlTerm) RqlTerm {
		return acc.Add(num)
	})
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"reduction": 135, "group": false},
		map[string]interface{}{"reduction": 70, "group": true},
	})
}

func (s *RethinkSuite) TestAggregationGroupMapReduceUngroup(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group(func(row RqlTerm) RqlTerm {
		return row.Field("id").Mod(2).Eq(0)
	}).Map(func(row RqlTerm) RqlTerm {
		return row.Field("num")
	}).Reduce(func(acc, num RqlTerm) RqlTerm {
		return acc.Add(num)
	}).Ungroup().OrderBy("reduction")
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"reduction": 70, "group": true},
		map[string]interface{}{"reduction": 135, "group": false},
	})
}

func (s *RethinkSuite) TestAggregationGroupMapReduceTable(c *test.C) {
	// Ensure table + database exist
	DbCreate("test").Exec(sess)
	Db("test").TableCreate("TestAggregationGroupedMapReduceTable").Exec(sess)

	// Insert rows
	err := Db("test").Table("TestAggregationGroupedMapReduceTable").Insert(objList).Exec(sess)
	c.Assert(err, test.IsNil)

	var response []interface{}
	query := Db("test").Table("TestAggregationGroupedMapReduceTable").Group(func(row RqlTerm) RqlTerm {
		return row.Field("id").Mod(2).Eq(0)
	}).Map(func(row RqlTerm) RqlTerm {
		return row.Field("num")
	}).Reduce(func(acc, num RqlTerm) RqlTerm {
		return acc.Add(num)
	})
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"reduction": 135, "group": false},
		map[string]interface{}{"reduction": 70, "group": true},
	})
}

func (s *RethinkSuite) TestAggregationGroupCount(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1")
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"group": 1, "reduction": []interface{}{
			map[string]interface{}{"id": 1, "num": 0, "g1": 1, "g2": 1},
			map[string]interface{}{"num": 15, "g1": 1, "g2": 1, "id": 6},
			map[string]interface{}{"id": 7, "num": 0, "g1": 1, "g2": 2},
		}},
		map[string]interface{}{"group": 2, "reduction": []interface{}{
			map[string]interface{}{"g1": 2, "g2": 2, "id": 2, "num": 5},
			map[string]interface{}{"num": 0, "g1": 2, "g2": 3, "id": 4},
			map[string]interface{}{"num": 100, "g1": 2, "g2": 3, "id": 5},
			map[string]interface{}{"g2": 3, "id": 9, "num": 25, "g1": 2},
		}},
		map[string]interface{}{"group": 3, "reduction": []interface{}{
			map[string]interface{}{"num": 10, "g1": 3, "g2": 2, "id": 3},
		}},
		map[string]interface{}{"group": 4, "reduction": []interface{}{
			map[string]interface{}{"id": 8, "num": 50, "g1": 4, "g2": 2},
		}},
	})
}

func (s *RethinkSuite) TestAggregationGroupSum(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Sum("num")
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"group": 1, "reduction": 15},
		map[string]interface{}{"reduction": 130, "group": 2},
		map[string]interface{}{"reduction": 10, "group": 3},
		map[string]interface{}{"group": 4, "reduction": 50},
	})
}

func (s *RethinkSuite) TestAggregationGroupAvg(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Avg("num")
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"reduction": 15, "group": 1},
		map[string]interface{}{"group": 2, "reduction": 130},
		map[string]interface{}{"group": 3, "reduction": 10},
		map[string]interface{}{"group": 4, "reduction": 50},
	})
}

func (s *RethinkSuite) TestAggregationGroupMin(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Min("num")
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"group": 1, "reduction": map[string]interface{}{"id": 1, "num": 0, "g1": 1, "g2": 1}},
		map[string]interface{}{"reduction": map[string]interface{}{"num": 0, "g1": 2, "g2": 3, "id": 4}, "group": 2},
		map[string]interface{}{"group": 3, "reduction": map[string]interface{}{"num": 10, "g1": 3, "g2": 2, "id": 3}},
		map[string]interface{}{"group": 4, "reduction": map[string]interface{}{"g2": 2, "id": 8, "num": 50, "g1": 4}},
	})
}

func (s *RethinkSuite) TestAggregationGroupMax(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Max("num")
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"reduction": map[string]interface{}{"num": 15, "g1": 1, "g2": 1, "id": 6}, "group": 1},
		map[string]interface{}{"group": 2, "reduction": map[string]interface{}{"num": 100, "g1": 2, "g2": 3, "id": 5}},
		map[string]interface{}{"group": 3, "reduction": map[string]interface{}{"num": 10, "g1": 3, "g2": 2, "id": 3}},
		map[string]interface{}{"group": 4, "reduction": map[string]interface{}{"g2": 2, "id": 8, "num": 50, "g1": 4}},
	})
}

func (s *RethinkSuite) TestAggregationMultipleGroupSum(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1", "g2").Sum("num")
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"group": []interface{}{1, 1}, "reduction": 15},
		map[string]interface{}{"reduction": 0, "group": []interface{}{1, 2}},
		map[string]interface{}{"group": []interface{}{2, 2}, "reduction": 5},
		map[string]interface{}{"reduction": 125, "group": []interface{}{2, 3}},
		map[string]interface{}{"group": []interface{}{3, 2}, "reduction": 10},
		map[string]interface{}{"group": []interface{}{4, 2}, "reduction": 50},
	})
}

func (s *RethinkSuite) TestAggregationGroupChained(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1").Max("num").Field("g2")
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"group": 1, "reduction": 1},
		map[string]interface{}{"group": 2, "reduction": 3},
		map[string]interface{}{"group": 3, "reduction": 2},
		map[string]interface{}{"group": 4, "reduction": 2},
	})
}

func (s *RethinkSuite) TestAggregationGroupUngroup(c *test.C) {
	var response []interface{}
	query := Expr(objList).Group("g1", "g2").Max("num").Ungroup()
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, JsonEquals, []interface{}{
		map[string]interface{}{"group": []interface{}{1, 1}, "reduction": map[string]interface{}{"g1": 1, "g2": 1, "id": 6, "num": 15}},
		map[string]interface{}{"group": []interface{}{1, 2}, "reduction": map[string]interface{}{"g1": 1, "g2": 2, "id": 7, "num": 0}},
		map[string]interface{}{"group": []interface{}{2, 2}, "reduction": map[string]interface{}{"g1": 2, "g2": 2, "id": 2, "num": 5}},
		map[string]interface{}{"group": []interface{}{2, 3}, "reduction": map[string]interface{}{"g1": 2, "g2": 3, "id": 5, "num": 100}},
		map[string]interface{}{"group": []interface{}{3, 2}, "reduction": map[string]interface{}{"g2": 2, "id": 3, "num": 10, "g1": 3}},
		map[string]interface{}{"reduction": map[string]interface{}{"num": 50, "g1": 4, "g2": 2, "id": 8}, "group": []interface{}{4, 2}},
	})
}

func (s *RethinkSuite) TestAggregationContains(c *test.C) {
	var response interface{}
	query := Expr(arr).Contains(2)
	r, err := query.RunRow(sess)
	c.Assert(err, test.IsNil)

	err = r.Scan(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, true)
}
