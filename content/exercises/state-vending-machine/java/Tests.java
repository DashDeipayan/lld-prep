public class Tests {
    static VendingMachine stocked() {
        VendingMachine vm = new VendingMachine();
        vm.stock(new Product("A1", "Cola", 125), 1);
        vm.stock(new Product("B2", "Chips", 50), 2);
        return vm;
    }

    static void testStartsSoldOutAndRejectsCoins() {
        VendingMachine vm = new VendingMachine();
        T.eq("SOLD_OUT", vm.state(), "new machine");
        T.throwsEx(IllegalStateException.class, () -> vm.insertCoin(25), "coin while sold out");
        T.throwsEx(IllegalStateException.class, () -> vm.select("A1"), "select while sold out");
        T.eq(0, vm.cancel(), "cancel while sold out");
    }

    static void testStockingMovesToIdle() {
        VendingMachine vm = stocked();
        T.eq("IDLE", vm.state(), "after stocking");
        T.eq(2, vm.quantity("B2"), "quantity");
        vm.stock(new Product("B2", "Chips", 50), 3);
        T.eq(5, vm.quantity("B2"), "stocking adds");
        T.throwsEx(IllegalArgumentException.class, () -> vm.stock(new Product("C3", "Gum", 10), 0), "qty must be positive");
    }

    static void testCoinsAndInvalidCoin() {
        VendingMachine vm = stocked();
        T.throwsEx(IllegalArgumentException.class, () -> vm.insertCoin(3), "invalid denomination");
        T.eq("IDLE", vm.state(), "invalid coin changes nothing");
        vm.insertCoin(25);
        vm.insertCoin(10);
        T.eq("HAS_MONEY", vm.state(), "has money");
        T.eq(35, vm.balance(), "balance");
    }

    static void testSelectInIdleRejected() {
        T.throwsEx(IllegalStateException.class, () -> stocked().select("A1"), "insert money first");
    }

    static void testSuccessfulSaleWithChange() {
        VendingMachine vm = stocked();
        vm.insertCoin(100);
        vm.insertCoin(25);
        vm.insertCoin(25);
        Dispense d = vm.select("A1");
        T.eq("A1", d.product().code(), "product dispensed");
        T.eq(25, d.change(), "change");
        T.eq(0, vm.balance(), "balance reset");
        T.eq(0, vm.quantity("A1"), "one unit taken");
        T.eq("IDLE", vm.state(), "stock remains (chips) -> IDLE");
    }

    static void testFailedSelectsKeepState() {
        VendingMachine vm = stocked();
        vm.insertCoin(25);
        T.throwsEx(IllegalStateException.class, () -> vm.select("A1"), "insufficient funds");
        T.throwsEx(IllegalArgumentException.class, () -> vm.select("ZZ"), "unknown product");
        T.eq("HAS_MONEY", vm.state(), "still has money");
        T.eq(25, vm.balance(), "balance untouched");
        T.throwsEx(IllegalStateException.class, () -> vm.stock(new Product("C3", "Gum", 10), 1), "cannot stock mid-transaction");
    }

    static void testOutOfStockAndSoldOut() {
        VendingMachine vm = new VendingMachine();
        vm.stock(new Product("A1", "Cola", 125), 1);
        vm.stock(new Product("B2", "Chips", 50), 1);
        vm.insertCoin(100);
        vm.insertCoin(25);
        vm.select("A1");
        vm.insertCoin(100);
        T.throwsEx(IllegalStateException.class, () -> vm.select("A1"), "A1 out of stock");
        Dispense d = vm.select("B2");
        T.eq(50, d.change(), "change from 100 for 50");
        T.eq("SOLD_OUT", vm.state(), "last unit sold -> SOLD_OUT");
    }

    static void testCancelRefunds() {
        VendingMachine vm = stocked();
        vm.insertCoin(100);
        vm.insertCoin(5);
        T.eq(105, vm.cancel(), "full refund");
        T.eq("IDLE", vm.state(), "back to idle");
        T.eq(0, vm.balance(), "balance cleared");
    }
}
