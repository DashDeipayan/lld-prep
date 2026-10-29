from lldtest import check, eq, raises
from solution import EqualSplit, ExactSplit, PercentSplit, SettlementError, Splitwise, Transfer


def sw(*users):
    s = Splitwise()
    for u in users:
        s.add_user(u)
    return s


# ---------------- step 1 ----------------
def s1_equal_split_distributes_remainder():
    eq([34, 33, 33], list(EqualSplit().shares(100, ["a", "b", "c"]).values()), "100 / 3")
    eq(["a", "b", "c"], list(EqualSplit().shares(100, ["a", "b", "c"]).keys()), "participant order")
    eq({"x": 7}, EqualSplit().shares(7, ["x"]), "single participant")


def s1_exact_split_validates():
    eq({"a": 70, "b": 30}, ExactSplit({"a": 70, "b": 30}).shares(100, ["a", "b"]), "exact")
    raises(ValueError, lambda: ExactSplit({"a": 70, "b": 20}).shares(100, ["a", "b"]), "sum mismatch")
    raises(ValueError, lambda: ExactSplit({"a": 100}).shares(100, ["a", "b"]), "missing participant")


def s1_percent_split_remainders():
    eq([4, 3, 3], list(PercentSplit({"a": 33, "b": 33, "c": 34}).shares(10, ["a", "b", "c"]).values()), "leftover goes in participant order")
    eq([33, 33, 34], list(PercentSplit({"a": 33, "b": 33, "c": 34}).shares(100, ["a", "b", "c"]).values()), "no leftover")
    raises(ValueError, lambda: PercentSplit({"a": 50, "b": 40}).shares(100, ["a", "b"]), "percents must sum to 100")
    raises(ValueError, lambda: EqualSplit().shares(0, ["a"]), "amount must be positive")
    raises(ValueError, lambda: EqualSplit().shares(10, []), "no participants")


# ---------------- step 2 ----------------
def s2_users_and_expenses():
    s = sw("a", "b", "c")
    raises(ValueError, lambda: s.add_user("a"), "duplicate user")
    eq("E1", s.add_expense("a", 300, ["a", "b", "c"], EqualSplit()), "expense id")
    eq(100, s.owes("b", "a"), "b owes a its share")
    eq(100, s.owes("c", "a"), "c owes a its share")
    eq(0, s.owes("a", "a"), "payer's own share is not a debt")
    eq("E2", s.add_expense("b", 50, ["c"], ExactSplit({"c": 50})), "payer need not participate")
    eq(50, s.owes("c", "b"), "c owes b")


def s2_expense_validation():
    s = sw("a", "b")
    raises(ValueError, lambda: s.add_expense("z", 10, ["a"], EqualSplit()), "unknown payer")
    raises(ValueError, lambda: s.add_expense("a", 10, ["a", "q"], EqualSplit()), "unknown participant")
    raises(ValueError, lambda: s.add_expense("a", 10, ["b", "b"], EqualSplit()), "duplicate participant")
    eq(0, s.owes("b", "a"), "failed expenses change nothing")


# ---------------- step 3 ----------------
def s3_netting_both_directions():
    s = sw("a", "b")
    s.add_expense("b", 30, ["a"], EqualSplit())
    s.add_expense("a", 10, ["b"], EqualSplit())
    eq(20, s.owes("a", "b"), "a owes b 30 - 10")
    eq(0, s.owes("b", "a"), "never negative")


def s3_balances_sum_to_zero():
    s = sw("a", "b", "c", "d")
    s.add_expense("a", 1000, ["a", "b", "c", "d"], EqualSplit())
    s.add_expense("c", 333, ["a", "b", "c"], EqualSplit())
    eq(750 - 111, s.balance("a"), "a: owed 750, owes c 111")
    eq(-250 - 111, s.balance("b"), "b owes a 250 and c 111")
    eq(0, sum(s.balance(u) for u in "abcd"), "money is conserved")


# ---------------- step 4 ----------------
def s4_settle_up():
    s = sw("a", "b")
    s.add_expense("a", 100, ["a", "b"], EqualSplit())
    s.settle("b", "a", 30)
    eq(20, s.owes("b", "a"), "partial settlement")
    raises(SettlementError, lambda: s.settle("b", "a", 21), "cannot overpay")
    raises(ValueError, lambda: s.settle("b", "a", 0), "positive amount")
    s.settle("b", "a", 20)
    eq(0, s.balance("a"), "fully settled")


# ---------------- step 5 ----------------
def s5_simplify_chain():
    s = sw("a", "b", "c")
    s.add_expense("b", 50, ["a"], ExactSplit({"a": 50}))
    s.add_expense("c", 50, ["b"], ExactSplit({"b": 50}))
    eq([Transfer("a", "c", 50)], s.simplify(), "a -> b -> c collapses to a -> c")
    eq(50, s.owes("a", "b"), "simplify does not change balances")


def s5_simplify_greedy_order_and_bound():
    s = sw("a", "b", "c", "d", "e")
    s.add_expense("a", 400, ["b", "c", "d", "e"], EqualSplit())
    s.add_expense("b", 300, ["c", "d", "e"], EqualSplit())
    # balances: a +400, b +300-100=+200, c -200, d -200, e -200
    plan = s.simplify()
    eq([Transfer("c", "a", 200), Transfer("d", "a", 200), Transfer("e", "b", 200)], plan,
       "largest creditor with largest debtor, ties by id")
    check(len(plan) <= 4, "at most n-1 transfers")
    eq([], sw("x", "y").simplify(), "nothing to settle")
