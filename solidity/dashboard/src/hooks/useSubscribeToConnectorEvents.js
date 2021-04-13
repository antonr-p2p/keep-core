import React, { useEffect } from "react"
import { useDispatch, useSelector } from "react-redux"
import { useWeb3Context } from "../components/WithWeb3Context"
import { WALLETS } from "../constants/constants"
import { useModal } from "./useModal"
import { WalletSelectionModal } from "../components/WalletSelectionModal"
import { useLocation, useHistory } from "react-router-dom"
import useWalletAddressFromUrl from "./useWalletAddressFromUrl"
import { Deferred } from "../contracts"

const useSubscribeToConnectorEvents = () => {
  const dispatch = useDispatch()
  const { isConnected, connector, yourAddress, web3 } = useWeb3Context()
  const { openModal } = useModal()
  const { transactionQueue } = useSelector((state) => state.transactions)
  const history = useHistory()
  const location = useLocation()
  const walletAddressFromUrl = useWalletAddressFromUrl()

  useEffect(() => {
    const accountChangedHandler = (address) => {
      dispatch({ type: "app/account_changed", payload: { address } })
    }

    const disconnectHandler = () => {
      dispatch({ type: "app/logout" })
    }

    const showChooseWalletModal = (payload) => {
      dispatch({
        type: "transactions/transaction_added_to_queue",
        payload: payload,
      })
      openModal(<WalletSelectionModal />, {
        title: "Select Wallet",
      })
    }

    const sendNextTransactionFromQueue = async (
      transactions,
      transactionIndex = 0
    ) => {
      web3.eth.sendTransaction(
        transactions[transactionIndex].params[0],
        async (error, response) => {
          if (!error) {
            const nextIndex = transactionIndex + 1
            if (transactions[nextIndex])
              await sendNextTransactionFromQueue(transactions, nextIndex)
          }
        }
      )
    }

    const executeTransactionsInQueue = async (transactions) => {
      if (transactions.length > 0) {
        dispatch({
          type: "transactions/clear_queue",
        })
        await sendNextTransactionFromQueue(transactions)
      }
    }

    if (isConnected && connector) {
      dispatch({ type: "app/login", payload: { address: yourAddress } })
      connector.on("accountsChanged", accountChangedHandler)
      connector.once("disconnect", disconnectHandler)

      if (connector.name === WALLETS.EXPLORER_MODE.name) {
        connector.eventEmitter.on(
          "chooseWalletAndSendTransaction",
          showChooseWalletModal
        )
      } else {
        executeTransactionsInQueue(transactionQueue)
        if (walletAddressFromUrl) {
          const newPath = location.pathname.replace(
            "/" + walletAddressFromUrl,
            ""
          )
          history.push({ pathname: newPath })
        }
      }
    }

    return () => {
      if (connector) {
        connector.removeListener("accountsChanged", accountChangedHandler)
        connector.removeListener("disconnect", disconnectHandler)
        if (connector.name === WALLETS.EXPLORER_MODE.name) {
          connector.eventEmitter.removeListener(
            "chooseWalletAndSendTransaction",
            showChooseWalletModal
          )
        }
      }
    }
  }, [isConnected, connector, dispatch, yourAddress, web3])
}

export default useSubscribeToConnectorEvents
