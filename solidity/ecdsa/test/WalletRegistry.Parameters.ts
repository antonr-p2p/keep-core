import { waffle, helpers } from "hardhat"
import { expect } from "chai"

import { walletRegistryFixture } from "./fixtures"

import type { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers"
import type { WalletRegistry, WalletRegistryStub } from "../typechain"

const { createSnapshot, restoreSnapshot } = helpers.snapshot

describe("WalletRegistry - Parameters", async () => {
  let walletRegistry: WalletRegistryStub & WalletRegistry

  let deployer: SignerWithAddress
  let walletOwner: SignerWithAddress
  let thirdParty: SignerWithAddress

  before("load test fixture", async () => {
    await createSnapshot()

    // eslint-disable-next-line @typescript-eslint/no-extra-semi
    ;({ walletRegistry, walletOwner, deployer, thirdParty } =
      await waffle.loadFixture(walletRegistryFixture))
  })

  after(async () => {
    await restoreSnapshot()
  })

  describe("updateDkgParameters", async () => {
    context("when called by the deployer", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry.connect(deployer).updateDkgParameters(1, 2, 3, 4)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by the wallet owner", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry.connect(walletOwner).updateDkgParameters(1, 2, 3, 4)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by a third party", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry.connect(thirdParty).updateDkgParameters(1, 2, 3, 4)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })
  })

  describe("updateRewardParameters", async () => {
    context("when called by the deployer", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry.connect(deployer).updateRewardParameters(1)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by the wallet owner", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry.connect(walletOwner).updateRewardParameters(1)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by a third party", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry.connect(thirdParty).updateRewardParameters(1)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })
  })

  describe("updateSlashingParameters", async () => {
    context("when called by the deployer", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry.connect(deployer).updateSlashingParameters(1)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by the wallet owner", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry.connect(walletOwner).updateSlashingParameters(1)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by a third party", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry.connect(thirdParty).updateSlashingParameters(1)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })
  })

  describe("updateWalletParameters", async () => {
    context("when called by the deployer", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry
            .connect(deployer)
            .updateWalletParameters(thirdParty.address)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by the wallet owner", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry
            .connect(walletOwner)
            .updateWalletParameters(thirdParty.address)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by a third party", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry
            .connect(thirdParty)
            .updateWalletParameters(thirdParty.address)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })
  })

  describe("updateRandomBeacon", async () => {
    context("when called by the deployer", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry
            .connect(deployer)
            .updateRandomBeacon(thirdParty.address)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by the wallet owner", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry
            .connect(walletOwner)
            .updateRandomBeacon(thirdParty.address)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })

    context("when called by a third party", async () => {
      it("should revert", async () => {
        await expect(
          walletRegistry
            .connect(thirdParty)
            .updateRandomBeacon(thirdParty.address)
        ).to.be.revertedWith("Ownable: caller is not the owner")
      })
    })
  })
})
