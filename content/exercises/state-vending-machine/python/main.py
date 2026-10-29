from solution import MachineStateError, Product, VendingMachine

vm = VendingMachine()
print("state:", vm.state())
vm.stock(Product("A1", "Cola", 125), 1)
vm.stock(Product("B2", "Chips", 50), 2)
vm.insert_coin(100)
vm.insert_coin(25)
vm.insert_coin(25)
print(vm.select("A1"), "-> state", vm.state())
try:
    vm.select("B2")
except MachineStateError as e:
    print("error:", e)
vm.insert_coin(10)
print("refund", vm.cancel(), ", state", vm.state())
