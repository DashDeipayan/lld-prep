public class Main {
    public static void main(String[] args) {
        VendingMachine vm = new VendingMachine();
        System.out.println("state: " + vm.state());
        vm.stock(new Product("A1", "Cola", 125), 1);
        vm.stock(new Product("B2", "Chips", 50), 2);
        vm.insertCoin(100);
        vm.insertCoin(25);
        vm.insertCoin(25);
        System.out.println(vm.select("A1") + " -> state " + vm.state());
        try {
            vm.select("B2");
        } catch (IllegalStateException e) {
            System.out.println("error: " + e.getMessage());
        }
        vm.insertCoin(10);
        System.out.println("refund " + vm.cancel() + ", state " + vm.state());
    }
}
