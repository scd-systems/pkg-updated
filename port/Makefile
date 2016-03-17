# New ports collection makefile for:   pkg-updated
# Date created:        10 March 2016
# Whom:                cdornig
#
# $FreeBSD$
#

PORTNAME=	pkg-updated
PORTVERSION=	0.2.9
CATEGORIES=	sysutils

MAINTAINER=	cd@scd-systems.net
COMMENT=	A pkg update daemon

LICENSE=	BSD3CLAUSE

BUILD_DEPENDS=	go>=1.5.0:${PORTSDIR}/lang/go noweb:/devel/noweb gmake:${PORTSDIR}/devel/gmake

USE_GITHUB=	yes
GH_ACCOUNT=	scd-systems

MAKEFILE=       BSDmakefile

# PORTDOCS=       INSTALL README THANKS
INFO=           pkg-updated
INSTALL_TARGET= all

# MAN1=		pkg-updated.1

.include <bsd.port.mk>
