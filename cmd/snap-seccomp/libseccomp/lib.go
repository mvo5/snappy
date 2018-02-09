// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package libseccomp

//#cgo CFLAGS: -D_FILE_OFFSET_BITS=64
//#cgo pkg-config: libseccomp
//#cgo LDFLAGS: -Wl,-Bstatic -lseccomp -Wl,-Bdynamic
//
//#include <asm/ioctls.h>
//#include <ctype.h>
//#include <errno.h>
//#include <linux/can.h>
//#include <linux/netlink.h>
//#include <sched.h>
//#include <search.h>
//#include <stdbool.h>
//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <sys/prctl.h>
//#include <sys/quota.h>
//#include <sys/resource.h>
//#include <sys/socket.h>
//#include <sys/stat.h>
//#include <sys/types.h>
//#include <sys/utsname.h>
//#include <termios.h>
//#include <unistd.h>
// //The XFS interface requires a 64 bit file system interface
// //but we don't want to leak this anywhere else if not globally
// //defined.
//#ifndef _FILE_OFFSET_BITS
//#define _FILE_OFFSET_BITS 64
//#include <xfs/xqm.h>
//#undef _FILE_OFFSET_BITS
//#else
//#include <xfs/xqm.h>
//#endif
//#include <seccomp.h>
//#include <linux/sched.h>
//#include <linux/seccomp.h>
//#include <arpa/inet.h>
//
//#ifndef AF_IB
//#define AF_IB 27
//#define PF_IB AF_IB
//#endif				// AF_IB
//
//#ifndef AF_MPLS
//#define AF_MPLS 28
//#define PF_MPLS AF_MPLS
//#endif				// AF_MPLS
//
//#ifndef PR_CAP_AMBIENT
//#define PR_CAP_AMBIENT 47
//#define PR_CAP_AMBIENT_IS_SET    1
//#define PR_CAP_AMBIENT_RAISE     2
//#define PR_CAP_AMBIENT_LOWER     3
//#define PR_CAP_AMBIENT_CLEAR_ALL 4
//#endif				// PR_CAP_AMBIENT
//
//#ifndef PR_SET_THP_DISABLE
//#define PR_SET_THP_DISABLE 41
//#endif				// PR_SET_THP_DISABLE
//#ifndef PR_GET_THP_DISABLE
//#define PR_GET_THP_DISABLE 42
//#endif				// PR_GET_THP_DISABLE
//
//#ifndef PR_MPX_ENABLE_MANAGEMENT
//#define PR_MPX_ENABLE_MANAGEMENT 43
//#endif
//
//#ifndef PR_MPX_DISABLE_MANAGEMENT
//#define PR_MPX_DISABLE_MANAGEMENT 44
//#endif
//
// //FIXME: ARCH_BAD is defined as ~0 in libseccomp internally, however
// //       this leads to a build failure on 14.04. the important part
// //       is that its an invalid id for libseccomp.
//
//#define ARCH_BAD 0x7FFFFFFF
//#ifndef SCMP_ARCH_AARCH64
//#define SCMP_ARCH_AARCH64 ARCH_BAD
//#endif
//
//#ifndef SCMP_ARCH_PPC
//#define SCMP_ARCH_PPC ARCH_BAD
//#endif
//
//#ifndef SCMP_ARCH_PPC64LE
//#define SCMP_ARCH_PPC64LE ARCH_BAD
//#endif
//
//#ifndef SCMP_ARCH_PPC64
//#define SCMP_ARCH_PPC64 ARCH_BAD
//#endif
//
//#ifndef SCMP_ARCH_S390X
//#define SCMP_ARCH_S390X ARCH_BAD
//#endif
//
//
//typedef struct seccomp_data kernel_seccomp_data;
//
//__u32 htot32(__u32 arch, __u32 val)
//{
//	if (arch & __AUDIT_ARCH_LE)
//		return htole32(val);
//	else
//		return htobe32(val);
//}
//
//__u64 htot64(__u32 arch, __u64 val)
//{
//	if (arch & __AUDIT_ARCH_LE)
//		return htole64(val);
//	else
//		return htobe64(val);
//}
//
import "C"

import (
	// FIXME: we want github.com/seccomp/libseccomp-golang but that
	// will not work with trusty because libseccomp-golang checks
	// for the seccomp version and errors if it find one < 2.2.0
	"github.com/mvo5/libseccomp-golang"
)

const (
	SECCOMP_RET_ALLOW = C.SECCOMP_RET_ALLOW
	SECCOMP_RET_KILL  = C.SECCOMP_RET_KILL

	SCMP_ARCH_X86_64  = C.SCMP_ARCH_X86_64
	SCMP_ARCH_AARCH64 = C.SCMP_ARCH_AARCH64
	SCMP_ARCH_ARM     = C.SCMP_ARCH_ARM
	SCMP_ARCH_PPC64   = C.SCMP_ARCH_PPC64
	SCMP_ARCH_PPC64LE = C.SCMP_ARCH_PPC64LE
	SCMP_ARCH_PPC     = C.SCMP_ARCH_PPC
	SCMP_ARCH_S390X   = C.SCMP_ARCH_S390X
	SCMP_ARCH_X86     = C.SCMP_ARCH_X86

	PF_UNIX                     = C.PF_UNIX
	PF_LOCAL                    = C.PF_LOCAL
	PF_INET                     = C.PF_INET
	PF_INET6                    = C.PF_INET6
	PF_IPX                      = C.PF_IPX
	PF_NETLINK                  = C.PF_NETLINK
	PF_X25                      = C.PF_X25
	PF_AX25                     = C.PF_AX25
	PF_ATMPVC                   = C.PF_ATMPVC
	PF_APPLETALK                = C.PF_APPLETALK
	PF_PACKET                   = C.PF_PACKET
	PF_ALG                      = C.PF_ALG
	PF_BRIDGE                   = C.PF_BRIDGE
	PF_NETROM                   = C.PF_NETROM
	PF_ROSE                     = C.PF_ROSE
	PF_NETBEUI                  = C.PF_NETBEUI
	PF_SECURITY                 = C.PF_SECURITY
	PF_KEY                      = C.PF_KEY
	PF_ASH                      = C.PF_ASH
	PF_ECONET                   = C.PF_ECONET
	PF_SNA                      = C.PF_SNA
	PF_IRDA                     = C.PF_IRDA
	PF_PPPOX                    = C.PF_PPPOX
	PF_WANPIPE                  = C.PF_WANPIPE
	PF_BLUETOOTH                = C.PF_BLUETOOTH
	PF_RDS                      = C.PF_RDS
	PF_LLC                      = C.PF_LLC
	PF_TIPC                     = C.PF_TIPC
	PF_IUCV                     = C.PF_IUCV
	PF_RXRPC                    = C.PF_RXRPC
	PF_ISDN                     = C.PF_ISDN
	PF_PHONET                   = C.PF_PHONET
	PF_IEEE802154               = C.PF_IEEE802154
	PF_CAIF                     = C.PF_CAIF
	AF_NFC                      = C.AF_NFC
	PF_NFC                      = C.PF_NFC
	AF_VSOCK                    = C.AF_VSOCK
	PF_VSOCK                    = C.PF_VSOCK
	AF_IB                       = C.AF_IB
	PF_IB                       = C.PF_IB
	AF_MPLS                     = C.AF_MPLS
	PF_MPLS                     = C.PF_MPLS
	PF_CAN                      = C.PF_CAN
	PR_CAP_AMBIENT              = C.PR_CAP_AMBIENT
	PR_CAP_AMBIENT_RAISE        = C.PR_CAP_AMBIENT_RAISE
	PR_CAP_AMBIENT_LOWER        = C.PR_CAP_AMBIENT_LOWER
	PR_CAP_AMBIENT_IS_SET       = C.PR_CAP_AMBIENT_IS_SET
	PR_CAP_AMBIENT_CLEAR_ALL    = C.PR_CAP_AMBIENT_CLEAR_ALL
	PR_CAPBSET_READ             = C.PR_CAPBSET_READ
	PR_CAPBSET_DROP             = C.PR_CAPBSET_DROP
	PR_SET_CHILD_SUBREAPER      = C.PR_SET_CHILD_SUBREAPER
	PR_GET_CHILD_SUBREAPER      = C.PR_GET_CHILD_SUBREAPER
	PR_SET_DUMPABLE             = C.PR_SET_DUMPABLE
	PR_GET_DUMPABLE             = C.PR_GET_DUMPABLE
	PR_SET_ENDIAN               = C.PR_SET_ENDIAN
	PR_GET_ENDIAN               = C.PR_GET_ENDIAN
	PR_SET_FPEMU                = C.PR_SET_FPEMU
	PR_GET_FPEMU                = C.PR_GET_FPEMU
	PR_SET_FPEXC                = C.PR_SET_FPEXC
	PR_GET_FPEXC                = C.PR_GET_FPEXC
	PR_SET_KEEPCAPS             = C.PR_SET_KEEPCAPS
	PR_GET_KEEPCAPS             = C.PR_GET_KEEPCAPS
	PR_MCE_KILL                 = C.PR_MCE_KILL
	PR_MCE_KILL_GET             = C.PR_MCE_KILL_GET
	PR_SET_MM                   = C.PR_SET_MM
	PR_SET_MM_START_CODE        = C.PR_SET_MM_START_CODE
	PR_SET_MM_END_CODE          = C.PR_SET_MM_END_CODE
	PR_SET_MM_START_DATA        = C.PR_SET_MM_START_DATA
	PR_SET_MM_END_DATA          = C.PR_SET_MM_END_DATA
	PR_SET_MM_START_STACK       = C.PR_SET_MM_START_STACK
	PR_SET_MM_START_BRK         = C.PR_SET_MM_START_BRK
	PR_SET_MM_BRK               = C.PR_SET_MM_BRK
	PR_SET_MM_ARG_START         = C.PR_SET_MM_ARG_START
	PR_SET_MM_ARG_END           = C.PR_SET_MM_ARG_END
	PR_SET_MM_ENV_START         = C.PR_SET_MM_ENV_START
	PR_SET_MM_ENV_END           = C.PR_SET_MM_ENV_END
	PR_SET_MM_AUXV              = C.PR_SET_MM_AUXV
	PR_SET_MM_EXE_FILE          = C.PR_SET_MM_EXE_FILE
	PR_MPX_ENABLE_MANAGEMENT    = C.PR_MPX_ENABLE_MANAGEMENT
	PR_MPX_DISABLE_MANAGEMENT   = C.PR_MPX_DISABLE_MANAGEMENT
	PR_SET_NAME                 = C.PR_SET_NAME
	PR_GET_NAME                 = C.PR_GET_NAME
	PR_SET_NO_NEW_PRIVS         = C.PR_SET_NO_NEW_PRIVS
	PR_GET_NO_NEW_PRIVS         = C.PR_GET_NO_NEW_PRIVS
	PR_SET_PDEATHSIG            = C.PR_SET_PDEATHSIG
	PR_GET_PDEATHSIG            = C.PR_GET_PDEATHSIG
	PR_SET_PTRACER              = C.PR_SET_PTRACER
	PR_SET_SECCOMP              = C.PR_SET_SECCOMP
	PR_GET_SECCOMP              = C.PR_GET_SECCOMP
	PR_SET_SECUREBITS           = C.PR_SET_SECUREBITS
	PR_GET_SECUREBITS           = C.PR_GET_SECUREBITS
	PR_SET_THP_DISABLE          = C.PR_SET_THP_DISABLE
	PR_TASK_PERF_EVENTS_DISABLE = C.PR_TASK_PERF_EVENTS_DISABLE
	PR_TASK_PERF_EVENTS_ENABLE  = C.PR_TASK_PERF_EVENTS_ENABLE
	PR_GET_THP_DISABLE          = C.PR_GET_THP_DISABLE
	PR_GET_TID_ADDRESS          = C.PR_GET_TID_ADDRESS
	PR_SET_TIMERSLACK           = C.PR_SET_TIMERSLACK
	PR_GET_TIMERSLACK           = C.PR_GET_TIMERSLACK
	PR_SET_TIMING               = C.PR_SET_TIMING
	PR_GET_TIMING               = C.PR_GET_TIMING
	PR_SET_TSC                  = C.PR_SET_TSC
	PR_GET_TSC                  = C.PR_GET_TSC
	PR_SET_UNALIGN              = C.PR_SET_UNALIGN
	PR_GET_UNALIGN              = C.PR_GET_UNALIGN
	Q_SYNC                      = C.Q_SYNC
	Q_QUOTAON                   = C.Q_QUOTAON
	Q_QUOTAOFF                  = C.Q_QUOTAOFF
	Q_GETFMT                    = C.Q_GETFMT
	Q_GETINFO                   = C.Q_GETINFO
	Q_SETINFO                   = C.Q_SETINFO
	Q_GETQUOTA                  = C.Q_GETQUOTA
	Q_SETQUOTA                  = C.Q_SETQUOTA
	Q_XQUOTAON                  = C.Q_XQUOTAON
	Q_XQUOTAOFF                 = C.Q_XQUOTAOFF
	Q_XGETQUOTA                 = C.Q_XGETQUOTA
	Q_XSETQLIM                  = C.Q_XSETQLIM
	Q_XGETQSTAT                 = C.Q_XGETQSTAT
	Q_XQUOTARM                  = C.Q_XQUOTARM
	NETLINK_SOCK_DIAG           = C.NETLINK_SOCK_DIAG
	NETLINK_RDMA                = C.NETLINK_RDMA
	NETLINK_CRYPTO              = C.NETLINK_CRYPTO
	NETLINK_INET_DIAG           = C.NETLINK_INET_DIAG
)

type KernelSeccompData C.kernel_seccomp_data

// important for unit testing
type SeccompData C.kernel_seccomp_data

func (sc *SeccompData) SetNr(nr seccomp.ScmpSyscall) {
	sc.nr = C.int(C.htot32(C.__u32(sc.arch), C.__u32(nr)))
}
func (sc *SeccompData) SetArch(arch uint32) {
	sc.arch = C.htot32(C.__u32(arch), C.__u32(arch))
}
func (sc *SeccompData) SetArgs(args [6]uint64) {
	for i := range args {
		sc.args[i] = C.htot64(sc.arch, C.__u64(args[i]))
	}
}

type ScmpArch seccomp.ScmpArch

var (
	ArchAMD64   = seccomp.ArchAMD64
	ArchARM64   = seccomp.ArchARM64
	ArchARM     = seccomp.ArchARM
	ArchX86     = seccomp.ArchX86
	ArchPPC     = seccomp.ArchPPC
	ArchPPC64   = seccomp.ArchPPC64
	ArchPPC64LE = seccomp.ArchPPC64LE
	ArchS390X   = seccomp.ArchS390X
)
