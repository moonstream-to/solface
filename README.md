# `solface`
`I<Anything>` - Generate Solidity interfaces to any contract

`solface` is a command-line tool which helps you generate Solidity interfaces from smart contract ABIs.

## Installing `solface`

You can install `solface` using:

```
go install github.com/bugout-dev/solface
```

## Using `solface`

It's as simple as:

```
$ solface -name IOwnableERC20 fixtures/abis/OwnableERC20.json

interface IOwnableERC20 {
        // structs

        // events
        event Approval(address owner, address spender, uint256 value);
        event OwnershipTransferred(address previousOwner, address newOwner);
        event Transfer(address from, address to, uint256 value);

        // functions
        function allowance(address owner, address spender) external view returns (uint256);
        function approve(address spender, uint256 amount) external nonpayable returns (bool);
        function balanceOf(address account) external view returns (uint256);
        function decimals() external view returns (uint8);
        function decreaseAllowance(address spender, uint256 subtractedValue) external nonpayable returns (bool);
        function increaseAllowance(address spender, uint256 addedValue) external nonpayable returns (bool);
        function mint(address account, uint256 amount) external nonpayable;
        function name() external view returns (string);
        function owner() external view returns (address);
        function renounceOwnership() external nonpayable;
        function symbol() external view returns (string);
        function totalSupply() external view returns (uint256);
        function transfer(address recipient, uint256 amount) external nonpayable returns (bool);
        function transferFrom(address sender, address recipient, uint256 amount) external nonpayable returns (bool);
        function transferOwnership(address newOwner) external nonpayable;

        // errors
}
```

You can also pipe ABIs into `solface`:

```
$ cat fixtures/abis/OwnableERC20.json | solface -name IOwnableERC20

interface IOwnableERC20 {
        // structs

        // events
        event Approval(address owner, address spender, uint256 value);
        event OwnershipTransferred(address previousOwner, address newOwner);
        event Transfer(address from, address to, uint256 value);

        // functions
        function allowance(address owner, address spender) external view returns (uint256);
        function approve(address spender, uint256 amount) external nonpayable returns (bool);
        function balanceOf(address account) external view returns (uint256);
        function decimals() external view returns (uint8);
        function decreaseAllowance(address spender, uint256 subtractedValue) external nonpayable returns (bool);
        function increaseAllowance(address spender, uint256 addedValue) external nonpayable returns (bool);
        function mint(address account, uint256 amount) external nonpayable;
        function name() external view returns (string);
        function owner() external view returns (address);
        function renounceOwnership() external nonpayable;
        function symbol() external view returns (string);
        function totalSupply() external view returns (uint256);
        function transfer(address recipient, uint256 amount) external nonpayable returns (bool);
        function transferFrom(address sender, address recipient, uint256 amount) external nonpayable returns (bool);
        function transferOwnership(address newOwner) external nonpayable;

        // errors
}
```

Enjoy!

## Contributing to `solface`

PRs welcome. Please use our GitHub issues to communicate with us: https://github.com/bugout-dev/solface/issues/new
