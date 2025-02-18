// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

// Import OpenZeppelin Contracts
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/token/common/ERC2981.sol";
import "@limitbreak/creator-token-contracts/contracts/utils/CreatorTokenBase.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract Zero is CreatorTokenBase, ERC2981, ERC721URIStorage, AccessControl {
    uint256 private _nextTokenId;
    bytes32 public constant MINT_ROLE = keccak256("MINT_ROLE");
    uint256 public constant MAX_SUPPLY = 10000;

    /**
     * @dev Initializes the contract by setting a `name`, `symbol`, `owner`, `maxSupply`, and default royalty information.
     * @param name_ The name of the NFT collection.
     * @param symbol_ The symbol of the NFT collection.
     */
    constructor(string memory name_, string memory symbol_)
        ERC721(name_, symbol_)
    {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(MINT_ROLE, msg.sender);
    }

    function _requireCallerIsContractOwner() internal view virtual override {
        _checkRole(DEFAULT_ADMIN_ROLE);
    }

    /**
     * @dev Mints a new NFT to the specified `player` with the given `tokenURI`.
     * @param player The address that will receive the minted NFT.
     * @param tokenURI The URI pointing to the metadata of the NFT.
     * @return tokenId The ID of the newly minted NFT.
     */
    function mint(address player, string memory tokenURI)
        public
        onlyRole(MINT_ROLE)
        returns (uint256)
    {
        require(
            _nextTokenId < MAX_SUPPLY,
            "Zero: Supply reached the max limit!"
        );
        uint256 tokenId = _nextTokenId++;
        _mint(player, tokenId);
        _setTokenURI(tokenId, tokenURI);
        return tokenId;
    }

    /**
     * @dev Returns the total number of NFTs minted.
     * @return The total supply of NFTs.
     */
    function totalSupply() public view returns (uint256) {
        return _nextTokenId;
    }

    /**
     * @dev Sets the default royalty information that all tokens will default to.
     * @param receiver The address to receive royalty payments.
     * @param feeNumerator The royalty fee in basis points (e.g., 500 = 5%).
     */
    function setDefaultRoyalty(address receiver, uint96 feeNumerator)
        external
        onlyRole(DEFAULT_ADMIN_ROLE)
    {
        _setDefaultRoyalty(receiver, feeNumerator);
    }

    /**
     * @dev Removes default royalty information.
     */
    function deleteDefaultRoyalty() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _deleteDefaultRoyalty();
    }

    /**
     * @dev Sets royalty information for a specific token ID, overriding the global default.
     * @param tokenId The ID of the NFT.
     * @param receiver The address to receive royalty payments.
     * @param feeNumerator The royalty fee in basis points (e.g., 500 = 5%).
     */
    function setTokenRoyalty(
        uint256 tokenId,
        address receiver,
        uint96 feeNumerator
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _setTokenRoyalty(tokenId, receiver, feeNumerator);
    }

    /**
     * @dev Resets royalty information for the token ID back to the global default.
     * @param tokenId The ID of the NFT.
     */
    function resetTokenRoyalty(uint256 tokenId)
        external
        onlyRole(DEFAULT_ADMIN_ROLE)
    {
        _resetTokenRoyalty(tokenId);
    }

    /**
     * @dev Overrides the `supportsInterface` method to include ERC2981 support.
     * @param interfaceId The interface identifier, as specified in ERC-165.
     * @return `true` if the contract implements the requested interface.
     */
    function supportsInterface(bytes4 interfaceId)
        public
        view
        override(ERC721URIStorage, ERC2981, AccessControl)
        returns (bool)
    {
        return
            interfaceId == type(ICreatorToken).interfaceId ||
            super.supportsInterface(interfaceId);
    }

    /**
     * @dev Transfers `tokenId` from its current owner to `to`, or alternatively mints (or burns) if the current owner
     * (or `to`) is the zero address. Returns the owner of the `tokenId` before the update.
     *
     * The `auth` argument is optional. If the value passed is non 0, then this function will check that
     * `auth` is either the owner of the token, or approved to operate on the token (by the owner).
     *
     * Emits a {Transfer} event.
     *
     * NOTE: If overriding this function in a way that tracks balances, see also {_increaseBalance}.
     */
    function _update(
        address to,
        uint256 tokenId,
        address auth
    ) internal override returns (address) {
        _validateBeforeTransfer(_ownerOf(tokenId), to, tokenId);
        return super._update(to, tokenId, auth);
    }

    /**
     * @dev Returns the list of NFT IDs owned by the given address.
     * @param owner The address of the NFT owner.
     * @return ownedNFTs An array of NFT IDs owned by the address.
     */
    function getNFTsByOwner(address owner)
        external
        view
        returns (uint256[] memory)
    {
        uint256 balance = balanceOf(owner);
        uint256[] memory ownedNFTs = new uint256[](balance);
        uint256 index = 0;

        for (uint256 tokenId = 0; tokenId < _nextTokenId; tokenId++) {
            if (ownerOf(tokenId) == owner) {
                ownedNFTs[index] = tokenId;
                index++;
            }
        }
        return ownedNFTs;
    }
}