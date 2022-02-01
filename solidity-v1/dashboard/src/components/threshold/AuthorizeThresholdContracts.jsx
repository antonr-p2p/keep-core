import React, { useCallback } from "react"
import AddressShortcut from "./../AddressShortcut"
import Button from "../Button"
import { DataTable, Column } from "../DataTable"
import { ViewAddressInBlockExplorer } from "../ViewInBlockExplorer"
import { KEEP } from "../../utils/token.utils"
import { shortenAddress } from "../../utils/general.utils"
import resourceTooltipProps from "../../constants/tooltips"
import * as Icons from "../Icons"
import ReactTooltip from "react-tooltip"

const AuthorizeThresholdContracts = ({
  data,
  onAuthorizeBtn,
  onStakeBtn,
  onSelectOperator,
  selectedOperator,
  filterDropdownOptions,
}) => {
  return (
    <section className="tile">
      <DataTable
        data={data}
        itemFieldId="operatorAddress"
        title="Authorize Contracts"
        subtitle="Below are the available operator contracts to authorize."
        withTooltip
        tooltipProps={resourceTooltipProps.authorize}
        noDataMessage="No contracts to authorize."
        withFilterDropdown
        filterDropdownProps={{
          options: filterDropdownOptions,
          onSelect: onSelectOperator,
          valuePropertyName: "operatorAddress",
          labelPropertyName: "operatorAddress",
          selectedItem: selectedOperator,
          noItemSelectedText: "All operators",
          renderOptionComponent: ({ operatorAddress }) => (
            <OperatorDropdownItem operatorAddress={operatorAddress} />
          ),
          selectedItemComponent: (
            <OperatorDropdownItem
              operatorAddress={selectedOperator.operatorAddress}
            />
          ),
          allItemsFilterText: "All Operators",
        }}
        centered
      >
        <Column
          header="operator"
          field="operatorAddress"
          renderContent={({ operatorAddress }) => (
            <AddressShortcut address={operatorAddress} />
          )}
        />
        <Column
          header="stake"
          field="stakeAmount"
          renderContent={({ stakeAmount, isFromGrant }) => {
            return (
              <>
                <div>{KEEP.displayAmountWithSymbol(stakeAmount)}</div>
                <div className={"text-grey-50"} style={{ fontSize: "14px" }}>
                  {isFromGrant ? "Grant Tokens" : "Wallet Tokens"}
                </div>
              </>
            )
          }}
        />
        <Column
          header="contract"
          field=""
          renderContent={({ contracts, operatorAddress }) => (
            <AuthorizeContractItem
              key={contracts[0].contractName}
              {...contracts[0]}
              operatorAddress={operatorAddress}
            />
          )}
        />
        <Column
          headerStyle={{ width: "20%", textAlign: "right" }}
          header="actions"
          tdStyles={{ textAlign: "right" }}
          field=""
          renderContent={({
            contracts,
            operatorAddress,
            authorizerAddress,
            beneficiaryAddress,
            isStakedToT,
            stakeAmount,
          }) => (
            <AuthorizeActions
              key={contracts[0].contractName}
              {...contracts[0]}
              isStakedToT={isStakedToT}
              operatorAddress={operatorAddress}
              authorizerAddress={authorizerAddress}
              beneficiaryAddress={beneficiaryAddress}
              stakeAmount={stakeAmount}
              onAuthorizeBtn={onAuthorizeBtn}
              onStakeBtn={onStakeBtn}
            />
          )}
        />
      </DataTable>
    </section>
  )
}

// const styles = {
//   tooltipContentWrapper: { textAlign: "left", minWidth: "15rem" },
// }

const AuthorizeContractItem = ({ contractName, operatorContractAddress }) => {
  return (
    <div className="flex row wrap space-between center">
      <div>
        <div className="text-big">{contractName}</div>
        <ViewAddressInBlockExplorer address={operatorContractAddress} />
      </div>
    </div>
  )
}

const AuthorizeActions = ({
  contractName,
  operatorAddress,
  authorizerAddress,
  beneficiaryAddress,
  stakeAmount,
  isAuthorized,
  onAuthorizeBtn,
  onStakeBtn,
}) => {
  const onAuthorize = useCallback(
    async (awaitingPromise) => {
      await onAuthorizeBtn(
        {
          operatorAddress,
          authorizerAddress,
          beneficiaryAddress,
          stakeAmount,
          contractName,
        },
        awaitingPromise
      )
    },
    [
      contractName,
      operatorAddress,
      authorizerAddress,
      beneficiaryAddress,
      stakeAmount,
      onAuthorizeBtn,
    ]
  )

  const onStake = useCallback(
    async (awaitingPromise) => {
      await onStakeBtn(
        {
          operatorAddress,
          authorizerAddress,
          beneficiaryAddress,
          stakeAmount,
          contractName,
          isAuthorized,
        },
        awaitingPromise
      )
    },
    [
      contractName,
      operatorAddress,
      authorizerAddress,
      beneficiaryAddress,
      stakeAmount,
      isAuthorized,
      onStakeBtn,
    ]
  )

  return isAuthorized ? (
    <Button
      onClick={onStake}
      className="btn btn-secondary btn-semi-sm"
      style={{ marginLeft: "auto" }}
    >
      <Icons.AlertFill
        data-tip
        data-for={`stake-tooltip-for-operator-${operatorAddress}`}
        className={"tooltip--button-corner"}
      />
      <ReactTooltip
        id={`stake-tooltip-for-operator-${operatorAddress}`}
        place="top"
        type="dark"
        effect={"solid"}
        className={"react-tooltip-base react-tooltip-base--arrow-right"}
        offset={{ left: "100%!important" }}
      >
        <span>
          The stake amount is not yet confirmed. Click “Stake” to confirm the
          stake amount. This stake is not staked on Threshold until it is
          confirmed.
        </span>
      </ReactTooltip>
      stake
    </Button>
  ) : (
    <Button
      onClick={onAuthorize}
      className="btn btn-secondary btn-semi-sm"
      style={{ marginLeft: "auto" }}
    >
      authorize and stake
    </Button>
  )
}

const OperatorDropdownItem = React.memo(({ operatorAddress }) => (
  <span key={operatorAddress} title={operatorAddress}>
    {shortenAddress(operatorAddress)}
  </span>
))

export default AuthorizeThresholdContracts
