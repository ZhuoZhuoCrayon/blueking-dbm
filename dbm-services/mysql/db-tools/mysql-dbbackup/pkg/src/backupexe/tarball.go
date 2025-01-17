package backupexe

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

	"dbm-services/mysql/db-tools/mysql-dbbackup/pkg/go-pubpkg/cmutil"
	"dbm-services/mysql/db-tools/mysql-dbbackup/pkg/src/common"
	"dbm-services/mysql/db-tools/mysql-dbbackup/pkg/src/dbareport"
	"dbm-services/mysql/db-tools/mysql-dbbackup/pkg/src/logger"
	"dbm-services/mysql/db-tools/mysql-dbbackup/pkg/src/parsecnf"
	"dbm-services/mysql/db-tools/mysql-dbbackup/pkg/util"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

// PackageFile package backup files
type PackageFile struct {
	srcDir     string
	dstDir     string
	dstTarFile string
	cnf        *parsecnf.Cnf
	resultInfo *dbareport.BackupResult
	indexFile  *IndexContent
}

// MappingPackage Package multiple backup files
// sort file list
// traverse file list
// create new tar_writer
// write file to tar package
// calculate the sums of file size, compare it with size limit
// create new tar_writer
// loop ...
// write last file to tar package
func (p *PackageFile) MappingPackage() error {
	logger.Log.Infof("Tarball Package: src dir %s, iolimit %d MB/s", p.srcDir, p.cnf.Public.IOLimitMBPerSec)
	// collect IndexContent info
	p.indexFile.Init(&p.cnf.Public, p.resultInfo)

	tarFileNum := 0
	dstTarName := fmt.Sprintf(`%s_%d.tar`, p.dstDir, tarFileNum)
	var tarSize uint64 = 0
	var tarUtil = util.TarWriter{IOLimitMB: p.cnf.Public.IOLimitMBPerSec}
	if err := tarUtil.New(dstTarName); err != nil {
		return err
	}
	defer tarUtil.Close()
	tarSizeMaxBytes := p.cnf.Public.TarSizeThreshold * 1024 * 1024
	// The files are walked in lexical order
	walkErr := filepath.Walk(p.srcDir, func(filename string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.Join(common.TargetName, strings.TrimPrefix(filename, p.srcDir))
		isFile, written, err := tarUtil.WriteTar(header, filename)
		if err != nil {
			return err
		} else if !isFile {
			return nil
		}

		p.indexFile.addFileContent(dstTarName, filename, written)
		if err = os.Remove(filename); err != nil {
			logger.Log.Error("failed to remove file while taring, err:", err)
		}

		tarSize += uint64(written)
		if tarSize >= tarSizeMaxBytes {
			logger.Log.Infof("need to tar file, accumulated tar size: %d bytes, dstFile: %s", tarSize, dstTarName)
			if err = tarUtil.Close(); err != nil {
				return err
			}
			p.indexFile.TotalFilesize += tarSize
			tarSize = 0
			tarFileNum++
			// new tarUtil object will be used for next loop
			dstTarName = fmt.Sprintf(`%s_%d.tar`, p.dstDir, tarFileNum)
			if err = tarUtil.New(dstTarName); err != nil {
				return err
			}
		}
		return nil
	})
	logger.Log.Infof("need to tar file, accumulated tar size: %d bytes, dstFile: %s", tarSize, dstTarName)
	p.indexFile.TotalFilesize += tarSize
	if walkErr != nil {
		logger.Log.Error("walk dir, err: ", walkErr)
		return walkErr
	}
	logger.Log.Infof("old srcDir removing io is limited to: %d MB/s", p.cnf.Public.IOLimitMBPerSec)
	if err := cmutil.TruncateDir(p.srcDir, p.cnf.Public.IOLimitMBPerSec); err != nil {
		// if err := os.RemoveAll(p.srcDir); err != nil {
		logger.Log.Error("failed to remove useless backup files")
		return err
	}

	p.indexFile.addPrivFile(p.dstDir)
	if err := p.indexFile.RecordIndexContent(&p.cnf.Public); err != nil {
		return err
	}
	return nil
}

// SplittingPackage Firstly, put all backup files into the tar file. Secondly, split the tar file to multiple parts
func (p *PackageFile) SplittingPackage() error {
	// collect IndexContent
	p.indexFile.Init(&p.cnf.Public, p.resultInfo)

	// tar srcDir to tar
	if err := p.tarballDir(); err != nil {
		return err
	}
	if fileSize := cmutil.GetFileSize(p.dstTarFile); fileSize >= 0 {
		p.indexFile.TotalFilesize = uint64(fileSize)
	} else {
		return errors.Errorf("fail to get file size for %s, got %d", p.dstTarFile, fileSize)
	}

	// split tar file to parts
	if err := p.splitTarFile(p.dstTarFile); err != nil {
		return err
	}

	p.indexFile.addPrivFile(p.dstDir)
	if err := p.indexFile.RecordIndexContent(&p.cnf.Public); err != nil {
		return err
	}
	return nil
}

// tarballDir tar srcDir to dstTarFile
// remove srcDir if success
func (p *PackageFile) tarballDir() error {
	logger.Log.Infof("Tarball Package: src dir %s, iolimit %d MB/s", p.srcDir, p.cnf.Public.IOLimitMBPerSec)
	var tarUtil = util.TarWriter{IOLimitMB: p.cnf.Public.IOLimitMBPerSec}
	if err := tarUtil.New(p.dstTarFile); err != nil {
		return err
	}
	defer tarUtil.Close()

	walkErr := filepath.Walk(p.srcDir, func(filename string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.Join(common.TargetName, strings.TrimPrefix(filename, p.srcDir))
		isFile, _, err := tarUtil.WriteTar(header, filename)
		if err != nil {
			return err
		} else if !isFile {
			return nil
		}
		// TODO limit io rate when removing
		if err = os.Remove(filename); err != nil {
			logger.Log.Error("failed to remove file while taring, err:", err)
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	if err := os.RemoveAll(p.srcDir); err != nil {
		return err
	}
	return nil
}

// splitTarFile split Tar file into multiple part_file
// update indexFile
func (p *PackageFile) splitTarFile(destFile string) error {
	splitSpeed := int64(300) // default: 300MB/s
	if p.cnf.PhysicalBackup.SplitSpeed != 0 {
		splitSpeed = p.cnf.PhysicalBackup.SplitSpeed
	}
	logger.Log.Infof("Splitting Package: Tar file %s with iolimit %d MB/s", p.dstTarFile, splitSpeed)
	fileInfo, err := os.Stat(destFile)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("stat %s, err :%v", destFile, err))
		return err
	}
	filePartSize := int64(p.cnf.Public.TarSizeThreshold) * 1024 * 1024 // MB to bytes
	partNum := int(math.Ceil(float64(fileInfo.Size()) / float64(filePartSize)))
	if partNum == 1 {
		tarFilename := filepath.Base(destFile)
		p.indexFile.addFileContent(tarFilename, tarFilename, fileInfo.Size())
		return nil
	}

	// num >=1
	fi, err := os.OpenFile(destFile, os.O_RDONLY, os.ModePerm)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("open file %s, err :%v", destFile, err))
		return err
	}
	defer fi.Close()

	paddingSize := len(cast.ToString(partNum))
	for i := 0; i < partNum; i++ {
		dstTarName := strings.TrimSuffix(destFile, ".tar")
		partTarName := fmt.Sprintf(`%s.part_%0*d`, dstTarName, paddingSize, i) // ReSplitPart
		destFileWriter, err := os.OpenFile(partTarName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		// io.Copy will record fi Seek Position
		if written, err := cmutil.IOLimitRateWithChunk(destFileWriter, fi, splitSpeed, filePartSize); err == nil {
			_ = destFileWriter.Close()
			p.indexFile.addFileContent(partTarName, partTarName, filePartSize)
		} else {
			_ = destFileWriter.Close()
			if err == io.EOF { // read end
				p.indexFile.addFileContent(partTarName, partTarName, written)
				break
			}
			return err
		}
	}
	// remove old tar File
	logger.Log.Infof("old tar removing io is limited to: %d MB/s", p.cnf.Public.IOLimitMBPerSec)
	if err := cmutil.TruncateFile(p.dstTarFile, p.cnf.Public.IOLimitMBPerSec); err != nil {
		return err
	}
	return nil
}

// PackageBackupFiles package backup files
// resultInfo 里面还只有 base 信息，没有文件信息
func PackageBackupFiles(cnf *parsecnf.Cnf, resultInfo *dbareport.BackupResult) error {
	targetDir := path.Join(cnf.Public.BackupDir, common.TargetName)
	var packageFile = &PackageFile{
		srcDir:     targetDir,
		dstDir:     targetDir,
		dstTarFile: targetDir + ".tar",
		cnf:        cnf,
		resultInfo: resultInfo,
		indexFile:  &IndexContent{},
	}
	logger.Log.Infof("BackupResult:%+v", resultInfo)

	// package files, and produce the index file at the same time
	if strings.ToLower(cnf.Public.BackupType) == "logical" {
		if err := packageFile.MappingPackage(); err != nil {
			return err
		}
	} else if strings.ToLower(cnf.Public.BackupType) == "physical" {
		if err := packageFile.SplittingPackage(); err != nil {
			return err
		}
	}

	return nil
}
