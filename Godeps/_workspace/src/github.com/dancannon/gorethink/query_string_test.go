package gorethink

import (
	test "launchpad.net/gocheck"
)

func (s *RethinkSuite) TestStringMatchSuccess(c *test.C) {
	query := Expr("id:0,name:mlucy,foo:bar").Match("name:(\\w+)").Field("groups").Nth(0).Field("str")

	var response string
	r, err := query.RunRow(sess)
	c.Assert(err, test.IsNil)

	err = r.Scan(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, "mlucy")
}

func (s *RethinkSuite) TestStringMatchFail(c *test.C) {
	query := Expr("id:0,foo:bar").Match("name:(\\w+)").Field("groups").Nth(0).Field("str")

	r, err := query.RunRow(sess)
	c.Assert(err, test.NotNil)
	c.Assert(r, test.IsNil)
}

func (s *RethinkSuite) TestStringSplit(c *test.C) {
	query := Expr("a,b,c").Split(",")

	var response []string
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.DeepEquals, []string{"a", "b", "c"})
}

func (s *RethinkSuite) TestStringSplitMax(c *test.C) {
	query := Expr("a,b,c").Split(",", 1)

	var response []string
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.DeepEquals, []string{"a", "b,c"})
}

func (s *RethinkSuite) TestStringSplitWhitespace(c *test.C) {
	query := Expr("a b c").Split()

	var response []string
	r, err := query.Run(sess)
	c.Assert(err, test.IsNil)

	err = r.ScanAll(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.DeepEquals, []string{"a", "b", "c"})
}

func (s *RethinkSuite) TestStringMatchUpcase(c *test.C) {
	query := Expr("tESt").Upcase()

	var response string
	r, err := query.RunRow(sess)
	c.Assert(err, test.IsNil)

	err = r.Scan(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, "TEST")
}

func (s *RethinkSuite) TestStringMatchDowncase(c *test.C) {
	query := Expr("tESt").Downcase()

	var response string
	r, err := query.RunRow(sess)
	c.Assert(err, test.IsNil)

	err = r.Scan(&response)

	c.Assert(err, test.IsNil)
	c.Assert(response, test.Equals, "test")
}
