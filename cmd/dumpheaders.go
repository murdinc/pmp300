package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/murdinc/pmp300/pkg/pmp300"
	"github.com/spf13/cobra"
)

var dumpHeadersCmd = &cobra.Command{
	Use:   "dump-headers",
	Short: "Dump internal and external memory headers for comparison",
	Long: `Connects to the PMP300, reads and prints the DirectoryHeader and other
critical metadata from both internal flash and (if present) an external SmartMedia card.
This is used for debugging filesystem structure differences.`,
	RunE: runDumpHeaders,
}

func init() {
	rootCmd.AddCommand(dumpHeadersCmd)
}

func runDumpHeaders(cmd *cobra.Command, args []string) error {
	// Initialize device, handle storage switching based on global externalFlag
	pmp, port, err := getInitializedPMPDevice()
	if err != nil {
		return err
	}
	defer port.Close()

	// --- Dump Internal Memory Header ---
	fmt.Println("\n--- Internal Flash Memory Structure ---")
	if err := pmp.SwitchStorage(pmp300.StorageInternal); err != nil {
		fmt.Printf("Error switching to internal storage: %v\n", err)
	} else {
		fmt.Println("Switched to Internal Flash.")
		internalDir, err := pmp.ReadDirectory()
		if internalDir != nil { // Always print if a directory struct was returned, even with errors
			printDirectoryDetails(pmp, internalDir)
		}
		if err != nil { // Then print the error if there was one
			fmt.Printf("Checksum validation failed for internal directory: %v\n", err)
		}
	}

	// --- Dump External Memory Header ---
	fmt.Println("\n--- External SmartMedia Card Structure ---")
	if err := pmp.SwitchStorage(pmp300.StorageExternal); err != nil {
		fmt.Printf("Error switching to external storage: %v\n", err)
	} else {
		fmt.Println("Switched to External SmartMedia.")
		if present, err := pmp.DetectExternalStorage(); err != nil || !present {
			fmt.Printf("External SmartMedia card not detected or unreadable: %v\n", err)
			fmt.Println("Please ensure a SmartMedia card is inserted and properly seated.")
		} else {
			externalDir, err := pmp.ReadDirectory()
			if externalDir != nil { // Always print if a directory struct was returned, even with errors
				printDirectoryDetails(pmp, externalDir)
			}
			if err != nil { // Then print the error if there was one
				fmt.Printf("Checksum validation failed for external directory: %v\n", err)
			}
		}
	}

	return nil
}

func printDirectoryDetails(pmp *pmp300.Device, dir *pmp300.Directory) {
	fmt.Printf("  Storage Type: %s\n", pmp.GetCurrentStorage().String())
	fmt.Printf("  EntryCount: %d\n", dir.Header.EntryCount)
	fmt.Printf("  BlocksAvailable: %d\n", dir.Header.BlocksAvailable)
	fmt.Printf("  BlocksUsed: %d\n", dir.Header.BlocksUsed)
	fmt.Printf("  BlocksRemaining: %d\n", dir.Header.BlocksRemaining)
	fmt.Printf("  BlocksBad: %d\n", dir.Header.BlocksBad)
	fmt.Printf("  TimeLastUpdate: %s (%d Unix)\n", time.Unix(int64(dir.Header.TimeLastUpdate), 0).Format("2006-01-02 15:04:05"), dir.Header.TimeLastUpdate)
	fmt.Printf("  Version: 0x%04X (%d)\n", dir.Header.Version, dir.Header.Version)
	fmt.Printf("  Checksum1: 0x%04X\n", dir.Header.Checksum1)
	fmt.Printf("  Checksum2: 0x%04X\n", dir.Header.Checksum2)
	fmt.Printf("  NotUsed2: %X\n", dir.Header.NotUsed2)
	fmt.Printf("  NotUsed3 (first 16 bytes): %X...\n", dir.Header.NotUsed3[:16])

	// Raw bytes of the entire DirectoryHeader
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, &dir.Header)
	fmt.Printf("  Raw Header (first 64 bytes): %X...\n", buf.Bytes()[:64])

	fmt.Printf("  Files (first 3 entries, or all if less than 3):\n")
	for i := 0; i < int(dir.Header.EntryCount) && i < 3; i++ {
		entry := dir.Entries[i]
		fmt.Printf("    [%d] Name: %s, Size: %d, Blocks: %d, Pos: %d\n", i+1, bytes.TrimRight(entry.Name[:], "\x00"), entry.Size, entry.BlockCount, entry.BlockPosition)
	}
	if dir.Header.EntryCount > 3 {
		fmt.Printf("    ... and %d more files.\n", dir.Header.EntryCount-3)
	}

	// Dump a small portion of BlockUsage and FAT
	fmt.Printf("  BlockUsage (first 16 bytes): %X...\n", dir.BlockUsage[:16])
	fmt.Printf("  FAT (first 16 entries): %v...\n", dir.FAT[:16])
}
