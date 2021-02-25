import { AbstractConnector } from "./abstract-connector"
import { WALLETS } from "../constants/constants"

class InjectedConnector extends AbstractConnector {
  constructor() {
    super()
    this.provider = window.ethereum
  }

  enable = async () => {
    if (!this.provider) {
      throw new Error("window.ethereum provider not found")
    }

    this.provider = window.provider
    // https://docs.metamask.io/guide/ethereum-provider.html#ethereum-autorefreshonnetworkchange
    this.provider.autoRefreshOnNetworkChange = false

    try {
      return await this.provider.request({ method: "eth_requestAccounts" })
    } catch (error) {
      if (error.code === 4001) {
        // EIP-1193 userRejectedRequest error
        // If this happens, the user rejected the connection request.
        console.error("The user rejected the connection request.")
        throw new Error("User rejected request")
      }
      throw error
    }
  }

  // TODO
  disconnect = async () => {
    if (window.ethereum) {
      window.ethereum.stop()
    }
  }

  // TODO
  getChainId = async () => {
    throw Error("Implement first")
  }

  // TODO
  getNetworkId = async () => {
    throw Error("Implement first")
  }

  get name() {
    return this.provider.isMetamask
      ? WALLETS.METAMASK.name
      : WALLETS.INJECTED_WALLET.name
  }
}

export { InjectedConnector }
