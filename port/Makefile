# $FreeBSD$

PORTNAME=		pkg-updated
PORTVERSION=		0.3.0
CATEGORIES=		sysutils

MAINTAINER=		cd@scd-systems.net
COMMENT=		Pkg update daemon

LICENSE=		BSD3CLAUSE
LICENSE_FILE=		${WRKSRC}/LICENSE

BUILD_DEPENDS=		go>=1.5.0:${PORTSDIR}/lang/go noweb:${PORTSDIR}/devel/noweb
USES=			gmake

USE_GITHUB=		yes
GH_ACCOUNT=		scd-systems

OPTIONS_DEFINE=		PDF
PDF_DESC=		Create and install the documentation file (needs tex)

PDF_BUILD_DEPENDS+=	pdflatex:${PORTSDIR}/print/tex-formats dvips:${PORTSDIR}/print/tex-dvipsk

do-build:
	${SETENV} ${BUILD_ENV} GOPATH=${WRKSRC} go install github.com/mattn/go-sqlite3;
	${GMAKE} -C ${WRKSRC} extract;
	${SETENV} ${BUILD_ENV} GOPATH=${WRKSRC} go build -o ${WRKSRC}/pkg-updated ${WRKSRC}/pkg-updated.go;
	.if ${PORT_OPTIONS:PDF}
		${GMAKE} -C ${WRKSRC} pdf;
	.endif

do-install:
	${CP} ${WRKSRC}/pkg-updated ${STAGEDIR}/${PREFIX}/bin/pkg-updated
	${STRIP_CMD} ${STAGEDIR}/${PREFIX}/bin/pkg-updated
	${MKDIR} ${STAGEDIR}/${PREFIX}/etc/pkg-updated
	${CP} ${WRKSRC}/pkg-updated.conf ${STAGEDIR}/${PREFIX}/etc/pkg-updated/pkg-updated.conf
	${CP} ${WRKSRC}/rc.d/pkg-updated ${STAGEDIR}/${PREFIX}/etc/rc.d/pkg-updated
	.if ${PORT_OPTIONS:PDF}
		${MKDIR} ${STAGEDIR}/${PREFIX}/share/doc/pkg-updated
		${CP} ${WRKSRC}/pkg-updated.pdf ${STAGEDIR}/${PREFIX}/share/doc/pkg-updated
	.endif

.include <bsd.port.mk>
