from lldtest import eq, raises
from solution import MachineStateError, Product, VendingMachine


def stocked():
    vm = VendingMachine()
    vm.stock(Product("A1", "Cola", 125), 1)
    vm.stock(Product("B2", "Chips", 50), 2)
    return vm


def test_starts_sold_out_and_rejects_coins():
    vm = VendingMachine()
    eq("SOLD_OUT", vm.state(), "new machine")
    raises(MachineStateError, lambda: vm.insert_coin(25), "coin while sold out")
    raises(MachineStateError, lambda: vm.select("A1"), "select while sold out")
    eq(0, vm.cancel(), "cancel while sold out")


def test_stocking_moves_to_idle():
    vm = stocked()
    eq("IDLE", vm.state(), "after stocking")
    eq(2, vm.quantity("B2"), "quantity")
    vm.stock(Product("B2", "Chips", 50), 3)
    eq(5, vm.quantity("B2"), "stocking adds")
    raises(ValueError, lambda: vm.stock(Product("C3", "Gum", 10), 0), "qty must be positive")


def test_coins_and_invalid_coin():
    vm = stocked()
    raises(ValueError, lambda: vm.insert_coin(3), "invalid denomination")
    eq("IDLE", vm.state(), "invalid coin changes nothing")
    vm.insert_coin(25)
    vm.insert_coin(10)
    eq("HAS_MONEY", vm.state(), "has money")
    eq(35, vm.balance(), "balance")


def test_select_in_idle_rejected():
    raises(MachineStateError, lambda: stocked().select("A1"), "insert money first")


def test_successful_sale_with_change():
    vm = stocked()
    vm.insert_coin(100)
    vm.insert_coin(25)
    vm.insert_coin(25)
    d = vm.select("A1")
    eq("A1", d.product.code, "product dispensed")
    eq(25, d.change, "change")
    eq(0, vm.balance(), "balance reset")
    eq(0, vm.quantity("A1"), "one unit taken")
    eq("IDLE", vm.state(), "stock remains (chips) -> IDLE")


def test_failed_selects_keep_state():
    vm = stocked()
    vm.insert_coin(25)
    raises(MachineStateError, lambda: vm.select("A1"), "insufficient funds")
    raises(ValueError, lambda: vm.select("ZZ"), "unknown product")
    eq("HAS_MONEY", vm.state(), "still has money")
    eq(25, vm.balance(), "balance untouched")
    raises(MachineStateError, lambda: vm.stock(Product("C3", "Gum", 10), 1), "cannot stock mid-transaction")


def test_out_of_stock_and_sold_out():
    vm = VendingMachine()
    vm.stock(Product("A1", "Cola", 125), 1)
    vm.stock(Product("B2", "Chips", 50), 1)
    vm.insert_coin(100)
    vm.insert_coin(25)
    vm.select("A1")
    vm.insert_coin(100)
    raises(MachineStateError, lambda: vm.select("A1"), "A1 out of stock")
    d = vm.select("B2")
    eq(50, d.change, "change from 100 for 50")
    eq("SOLD_OUT", vm.state(), "last unit sold -> SOLD_OUT")


def test_cancel_refunds():
    vm = stocked()
    vm.insert_coin(100)
    vm.insert_coin(5)
    eq(105, vm.cancel(), "full refund")
    eq("IDLE", vm.state(), "back to idle")
    eq(0, vm.balance(), "balance cleared")
