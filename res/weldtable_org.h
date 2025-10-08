/*-------------------------------------------------------------------------------------------*/
/*  [�T�v]                                                                                   */
/*      �n�ڏ����e�[�u���e�k�`�r�g�������������݃v���O�����m�w�b�_���n                       */
/*                                                                                           */
/*                                     ------------------------------------------------------*/
/*                                    |    DATE    : CONTENT                                 */
/*                                     ------------------------------------------------------*/
/*                                    | 2016.01.22 : Original                                */
/*                                    | 2016.02.08 : �������݃^�C�~���O�����~�M���ɕύX  */
/*                                    | 2016.03.03 : �e�[�u�����e�ύX���ݒ蕔�ȑf��          */
/*                                    | 2025.06.26 : NR1�p�ɕύX                             */
/*                                    | 2025.09.08 : �����@�p�e�[�u���ǉ�(+4)                */
/*                                    | 2025.09.09 : �����@�p�e�[�u���ǉ�(+4)                */
/*                                    |            :                                         */
/*-------------------------------------------------------------------------------------------*/
#include <typedef.h>

//----------------------------------------------------------------------------
// �n�ڏ����e�[�u���̘A��
//  ���������ݍ�Ƃ̘A�ԁi�P�`�R�j
//    �P��ځ��P �^ �Q��ځ��Q �^ �R��ځ��R
//
#define		FLASH_WRITE_NO		1
//----------------------------------------------------------------------------

//----------------------------------------------------------------------------
// �n�ڏ����e�[�u���̏������݌�(�G���h�t���O�e�[�u���܂܂Ȃ��j
//  ���������ݍs���̖����s����̏������݌��́u�O�v�ɂ��邱�ƁB
//
//   ex.1
//      �n�ڏ����e�[�u���̏������݌� = 10
//
//       #define		WeldTableNum1st		10		//
//       #define		WeldTableNum2nd		0		// �� �u0�v�ɂ��邱��
//       #define		WeldTableNum3rd		0		// �� �u0�v�ɂ��邱��
//
//   ex.2
//      �n�ڏ����e�[�u���̏������݌� = 50
//
//       #define		WeldTableNum1st		37		// �P�H���ɍő�R�V�̃e�[�u���̏������݂��\
//       #define		WeldTableNum2nd		13		//
//       #define		WeldTableNum3rd		0		// �� �u0�v�ɂ��邱��
//
//
#define		WeldTableNum1st		31									// �P��ځF�n�ڏ����e�[�u������(MAX:37)
#define		WeldTableNum2nd		0									// �Q��ځF�n�ڏ����e�[�u������(MAX:37)
#define		WeldTableNum3rd		0									// �R��ځF�n�ڏ����e�[�u������(MAX:37)
//----------------------------------------------------------------------------

//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><> �萔���������� <><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>

//-----------------------------------------------------
// �������݊J�n�ʒu�̎�������
#define		WeldTableStart1st  	0									// �P��ځF�n�ڏ����e�[�u���������݊J�n�ʒu�i�n�ڏ����e�[�u���P�ʁj�O���擪�ʒu��FLASH�����L��
#define		WeldTableStart2nd  	WeldTableNum1st						// �Q��ځF�n�ڏ����e�[�u���������݊J�n�ʒu�i�n�ڏ����e�[�u���P�ʁF�{�P��ځj
#define		WeldTableStart3rd  	WeldTableNum1st+WeldTableNum2nd		// �R��ځF�n�ڏ����e�[�u���������݊J�n�ʒu�i�n�ڏ����e�[�u���P�ʁG�{�P��ځ{�Q��ځj
//-----------------------------------------------------
// �������ݍ�Ƃ̘A�Ԏ��ʎq�̎�������
#if (FLASH_WRITE_NO == 1)
	#define		FLASH_1ST			// �P��ڂ̏�������
#elif (FLASH_WRITE_NO == 2)
	#define		FLASH_2ND			// �Q��ڂ̏�������
#elif (FLASH_WRITE_NO == 3)
	#define		FLASH_3RD			// �R��ڂ̏�������
#else
	#define		FLASH_1ST			// �P��ڂ̏�������
#endif
//-----------------------------------------------------

#ifdef	FLASH_1ST			// �P���
	#if (WeldTableNum1st < 35)	// �������ޗn�ڏ��������R�S�ȉ��H
		#define		WELDTABLE_ENDCODE
	#else
		#if (WeldTableNum2nd == 0)	// ����̏������݂͖����H
			#define		WELDTABLE_ENDCODE
		#endif
	#endif
	#ifdef		WELDTABLE_ENDCODE	// �G���h�R�[�h�L��
		#define		WeldTableNum		WeldTableNum1st+1	// �n�ڏ����e�[�u�������i�G���h�R�[�h������Ί܂ނ��Ɓj
	#else
		#define		WeldTableNum		WeldTableNum1st		// �n�ڏ����e�[�u������
	#endif
	#define		WeldTableStart  	WeldTableStart1st	// �n�ڏ����e�[�u���������݊J�n�ʒu�i�n�ڏ����e�[�u���P�ʁj�O���擪�ʒu��FLASH�����L��
#elif defined(FLASH_2ND)	// �Q���
	#if (WeldTableNum2st < 37)	// �������ޗn�ڏ��������R�V�ȉ��H
		#define		WELDTABLE_ENDCODE
	#else
		#if (WeldTableNum3nd == 0)	// ����̏������݂͖����H
			#define		WELDTABLE_ENDCODE
		#endif
	#endif
	#ifdef		WELDTABLE_ENDCODE	// �G���h�R�[�h�L��
		#define		WeldTableNum		WeldTableNum2nd+1	// �n�ڏ����e�[�u�������i�G���h�R�[�h������Ί܂ނ��Ɓj
	#else
		#define		WeldTableNum		WeldTableNum2nd		// �n�ڏ����e�[�u������
	#endif
	#define		WeldTableStart  	WeldTableStart2nd	// �n�ڏ����e�[�u���������݊J�n�ʒu�i�n�ڏ����e�[�u���P�ʁj�O���擪�ʒu��FLASH�����L��
#else						// �R���
	#define		WELDTABLE_ENDCODE						// �R��ڂ̏������݂�����ꍇ�͕K���G���h�R�[�h�L��Ƃ���B
	#define		WeldTableNum		WeldTableNum3rd+1	// �n�ڏ����e�[�u�������i�G���h�R�[�h������Ί܂ނ��Ɓj
	#define		WeldTableStart  	WeldTableStart3rdd	// �n�ڏ����e�[�u���������݊J�n�ʒu�i�n�ڏ����e�[�u���P�ʁj�O���擪�ʒu��FLASH�����L��
#endif
//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>
//<><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><><>




/*--------------------------------------------------------------------------------------------------------------*/
/*	2025.03.14	:���쐬(BASE:VP1�p�n�ڏ����e�[�u��)											:	�J��			*/
/*	2025.03.27	:�σp�����[�^�ő區�ڐ����`(MAX:160)									:	�J��			*/
/*	2025.03.27	:���쐬(�σp�����[�^�̃e�[�u���������X�g�e�[�u���̊g���ɔ����ύX)			:	�J��			*/
/*	2025.04.11	:350NR1,500NR1�����ŏI�d�l�Ή�												:	�J��			*/
/*	2025.04.11	:�u_NR1�v���ꎞ�I�Ɂu_GX3�v�ɕύX�i�r���h�΍�j								:	�J��			*/
/*	2025.04.11	:�n�ڎ�ʃR�[�h���ꎞ�I�ɋ��d�l�ɕύX�i�r���h�΍�j							:	�J��			*/
/*	    .  .  	:																			:					*/
/*--------------------------------------------------------------------------------------------------------------*/
/*	[�T�v]																										*/
/*				�n�ڎ�ʖ��̗n�ڏ����e�[�u���i�m�q�P�p�j														*/
/*--------------------------------------------------------------------------------------------------------------*/
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�n�ڎ�ʃR�[�h															*/
/*			NOTE�F	�u�o1�p�𗬗p�ł���H						�F	2025.03.14	*/
/*				�F	@-1211075_00 �r�b�g�t�B�[���h				�F	----.--.--	*/
/*				�F	NR1�ŏI�d�l�p�ɍ��ڕύX						�F	2025.04.11	*/
/*------------------------------------------------------------------------------*/
/*	[�f�[�^����]�@���e�[�u���e�ʁF�Q�S�o�C�g									*/
/*		uchar		material				�ގ�								*/
/*		uchar		method					�n�ږ@								*/
/*		uchar		pulseCode				�p���X���[�h						*/
/*		uchar		pulseType				�p���X�^�C�v						*/
/*		uchar		wire					���C���a							*/
/*		uchar		extension				�˂��o����							*/
/*		uchar		tip						�`�b�v�^�C�v	0:�W�� 1:�گ�		*/	// @-R070901RK1  dummy1 �� tip
/*		uchar		Flag2					�n�ڎ�ʃt���O�Q	D0:�׸�ڽ		*/	// @-R091001RK1  dummy2 �� Flag2
/*																D1:���Đ���		*/
/*		uchar		Version					�o�[�W����							*/
/*		uchar		StandardFlag			�e�[�u���W���l�t���O				*/
/*		uchar		Flag3					�n�ڎ�ʃt���O�R					*/
/*		uchar		LowSputter				��X�p�b�^�t���O					*/	// [2025.04.11]
/*		uchar		rsv_1					�\��								*/
/*					  |															*/
/*		uchar		rsv_8					�\��								*/
/*		uchar		DPS_Lower				�c�o�r�ԍ��i���ʁj					*/	// [2025.04.11]
/*		uchar		DPS_Upper				�c�o�r�ԍ��i��ʁj					*/	// [2025.04.11]
/*		uchar		VER_Lower				�o�[�W�����ԍ��i���ʁj				*/	// [2025.04.11]
/*		uchar		VER_Upper				�o�[�W�����ԍ��i��ʁj				*/	// [2025.04.11]
/*------------------------------------------------------------------------------*/
typedef	struct
{
#if 0	//350NR1,500NR1		+++++[2025.04.11]+++++		���d�l��[2025.04.11]
	uchar		material		:	8;		//	+0	�ގ�
	uchar		method			:	8;		//	+1	�n�ږ@
	uchar		pulseMode		:	8;		//	+2	�p���X���[�h
	uchar		pulseType		:	8;		//	+3	�p���X�^�C�v
	uchar		wire			:	8;		//	+4	���C���a
	uchar		extension		:	8;		//	+5	�˂��o����
	uchar		tip				:	8;		//	+6	�`�b�v�^�C�v
	uchar		Flag2			:	8;		//	+7	�n�ڎ�ʃt���O�Q
	uchar		Version			:	8;		//	+8	�o�[�W����
	uchar		StandardFlag	:	8;		//	+9	�������o�[�W�����W���l�t���O
	uchar		Flag3			:	8;		//	+10	�n�ڎ�ʃt���O�R
	uchar		LowSputter		:	8;		//	+11	��X�p�b�^�t���O					[2025.04.11]
	uchar		rsv_1			:	8;		//	+12	�\��
	uchar		rsv_2			:	8;		//	+13	�\��
	uchar		rsv_3			:	8;		//	+14	�\��
	uchar		rsv_4			:	8;		//	+15	�\��
	uchar		rsv_5			:	8;		//	+16	�\��
	uchar		rsv_6			:	8;		//	+17	�\��
	uchar		rsv_7			:	8;		//	+18	�\��
	uchar		rsv_8			:	8;		//	+19	�\��
	uchar		DPS_Lower		:	8;		//	+20	�c�o�r�ԍ��i���ʁj					[2025.04.11]
	uchar		DPS_Upper		:	8;		//	+21	�c�o�r�ԍ��i��ʁj					[2025.04.11]
	uchar		VER_Lower		:	8;		//	+22	�o�[�W�����ԍ��i���ʁj				[2025.04.11]
	uchar		VER_Upper		:	8;		//	+23	�o�[�W�����ԍ��i��ʁj				[2025.04.11]
#else	//350NR1,500NR1		+++++[2025.04.11]-----
	uchar		material		:	8;		//	�ގ�
	uchar		method			:	8;		//	�n�ږ@
	uchar		pulseMode		:	8;		//	�p���X���[�h
	uchar		pulseType		:	8;		//	�p���X�^�C�v
	uchar		wire			:	8;		//	���C���a
	uchar		extension		:	8;		//	�˂��o����
	uchar		tip				:	8;		//	�`�b�v�^�C�v
	uchar		Flag2			:	8;		//	�n�ڎ�ʃt���O�Q
	uchar		Version			:	8;		//	�o�[�W����
	uchar		StandardFlag	:	8;		//	�������o�[�W�����W���l�t���O
	uchar		Flag3			:	8;		//	�n�ڎ�ʃt���O�R
	uchar		rsv_1			:	8;		//	�\��
	uchar		rsv_2			:	8;		//	�\��
	uchar		rsv_3			:	8;		//	�\��
	uchar		rsv_4			:	8;		//	�\��
	uchar		rsv_5			:	8;		//	�\��
	uchar		rsv_6			:	8;		//	�\��
	uchar		rsv_7			:	8;		//	�\��
	uchar		rsv_8			:	8;		//	�\��
	uchar		rsv_9			:	8;		//	�\��
	uchar		rsv_10			:	8;		//	�\��
	uchar		rsv_11			:	8;		//	�\��
	uchar		rsv_12			:	8;		//	�\��
	uchar		rsv_13			:	8;		//	�\��
#endif	//350NR1,500NR1		+++++[2025.04.11]-----
} WELDCODE, *PWELDCODE;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		���C���������x�e�[�u��													*/
/*			NOTE�F	�ő�T�P�Q��(1022A��)						�F	2025.03.14	*/
/*				�F	�ő�Q�T�U��(510A��)						�F	2025.04.11	*/
/*------------------------------------------------------------------------------*/
/*	[�f�[�^����]																*/
/*		usint		Speed[256]				�����e�[�u��( 0.001 m/min )			*/
/*												[ xx ]	xx�́A2A�P��			*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	usint		Speed[256];					//	�����e�[�u��(0.001m/min)[xx]	��xx�́A2A�P��			[512]->[256]	//[2025.04.11]
} A2STBL, *PA2STBL;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�ꌳ�d���e�[�u��														*/
/*			NOTE�F	�u�o1�p�𗬗p�ł���H						�F	2025.03.14	*/
/*------------------------------------------------------------------------------*/
/*	[�f�[�^����]																*/
/*		usint		Volt[256]				�ꌳ�d���e�[�u��( 0.1 V )			*/
/*												[ xx ]	xx�́A0.1m/min�P��		*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	usint		Volt[256];					//	�ꌳ�d���e�[�u��(0.1V)[xx]		��xx�́A0.1m/min�P��
} S2VTBL, *PS2VTBL;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		���Œ�p�����[�^�e�[�u��												*/
/*			NOTE�F	�u�o1�p�𗬗p�ł���H						�F	2025.03.14	*/
/*				�F	���ڐ����T�O�S���ڂɊg��(H379-H504��ǉ�)	�F	2025.04.11	*/
/*------------------------------------------------------------------------------*/
/*	[�f�[�^����]																*/
/*		usint		Parm[512]					�p�����[�^�i��Βl�j			*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	usint		Parm[504];					//	�p�����[�^�i��Βl�j			[378]->[504]	//[2025.04.11]
} WELDPARM, *PWELDPARM;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�σp���[���[�^�W���e�[�u��											*/
/*			NOTE�F	�u�o1�p�𗬗p�ł���H						�F	2025.03.14	*/
/*------------------------------------------------------------------------------*/
/*	[�f�[�^����]																*/
/*		float		a						�W����								*/
/*		float		b						�W����								*/
/*		float		c						�W����								*/
/*		float		min						�W��������							*/
/*		float       max						�W��������							*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	float		a;							//	�W����
	float		b;							//	�W����
	float		c;							//	�W����
	float		min;						//	�W��������
	float		max;						//	�W��������
} DCCALPARM, *PDCCALPARM;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�p�����[�^�e�[�u���i�p���X/�Z���j										*/
/*			NOTE�F	�u�o1�p�𗬗p�ł���H						�F	2025.03.14	*/
/*				�F	@-R040701RTW								�F	----.--.--	*/
/*				�F	@-1211075_02								�F	----.--.--	*/
/*				�F	VP1 AWC										�F	2016.02.05	*/
/*				�F	VP1 NEW_AWC									�F	2016.02.10	*/
/*				�F	VP1 NEW_AWC(�ڐA)							�F	2016.02.23	*/
/*------------------------------------------------------------------------------*/
/*	[�f�[�^����]																*/
/*		usint		SlowDown				�X���[�_�E��( 0.001 m/min )			*/
/*		usint		acc1					���������x�P( 0.1 )					*/
/*												�g�[�`�n�m�p					*/
/*		usint		acc2					�����x�Q							*/
/*												�����Z���p						*/
/*		usint		acc3					�����x�R							*/
/*												�����ύX�p						*/
/*		usint		acc4					�����x�S							*/
/*												�N���[�^�p						*/
/*		usint		acc5					�����x�T							*/
/*												�g�[�`�n�e�e�p					*/
/*		usint		Delay					�����f�B���C						*/
/*		usint		ContArc					�A���A�[�N�X�^�[�g����				*/
/*		usint		TigAcc					�����x								*/ // @-R060506RK1
/*												�s�h�f�t�B���[					*/ // @-R060506RK1
/*												�� �X���O���X�W���Ƃ��Ă��g�p	*/ // @-R091001RK1
/*		usint		TigDcc					�����x								*/ // @-R060506RK1
/*												�s�h�f�t�B���[					*/ // @-R060506RK1
/*		usint		SlowDown_Act			�A�N�e�B�u�F�X���[�_�E��			*/ // @-R091001RK1
/*		usint		acc2_Act				�A�N�e�B�u�F�����x�Q				*/ // @-R091001RK1
/*		usint		Delay_Act				�A�N�e�B�u�F�����f�B���C			*/ // @-R091001RK1
/*								---�`�v�b�֘A---								*/
/*		usint		awc_select				�w�ߒl����L���I��					*/
/*											�i�O�F�����^�P�F�L��j				*/
/*		usint		init_curr_percent		���������d���i���j					*/
/*		usint		init_volt_adj			���������d�������l					*/
/*											���ꌳ�d���ɑ΂��钲���l			*/
/*											�i�f�[�^�~0.1V	���f�[�^�͂Q���݁j	*/
/*		sint		all_init_time			�����������ԁi�����j				*/
/*		usint		init_up_time			���������X�^�[�g�X���[�v���ԁi�����j*/
/*		usint		init_dw_time			���������_�E���X���[�v���ԁi�����j	*/
/*		usint		init_limit_curr			���������d������					*/
/*											�i�f�[�^�~�P�`�@���f�[�^�͂Q���݁j	*/
/*		usint		crat_curr_percent		�N���[�^�����d���i���j				*/
/*		usint		crat_volt_adj			�N���[�^�d�������l					*/
/*											���ꌳ�d���ɑ΂��钲���l			*/
/*											�i�f�[�^�~0.1V�@���f�[�^�͂Q���݁j	*/
/*		usint		all_crat_time			�N���[�^�������ԁi�����j			*/
/*		usint		crat_dw_time			�N���[�^�����_�E���X���[�v����(ms)	*/
/*		usint		crat_skip_time			�N���[�^�X�L�b�v����				*/
/*							--- �V�`�v�b�֘A---									*/
/*		float		init_up_a				���������X�^�[�g�X���[�v�@��		*/
/*		float		init_up_b				���������X�^�[�g�X���[�v�@��		*/
/*		float		init_up_c				���������X�^�[�g�X���[�v�@��		*/
/*		usint		init_up_min				���������X�^�[�g�X���[�v�@������	*/
/*		usint		init_up_max				���������X�^�[�g�X���[�v�@������	*/
/*																				*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	usint		SlowDown;				//	�X���[�_�E��(0.001m/min)
	usint		acc1;					//	���������x�P(0.1)	�g�[�`�n�m�p
	usint		acc2;					//		�����x�Q		�����Z���p
	usint		acc3;					//		�����x�R		�����ύX�p
	usint		acc4;					//		�����x�S		�N���[�^�p
	usint		acc5;					//		�����x�T		�g�[�`�n�e�e�p
	usint		Delay;					//	�����f�B���C
	usint		ContArc;				//	�A���A�[�N�X�^�[�g����
	usint		TigAcc;					//	�����x	�s�h�f�t�B���[	���X���O���X�W���Ƃ��Ă��g�p
	usint		TigDcc;					//	�����x	�s�h�f�t�B���[
	usint		SlowDown_Act;			//	�A�N�e�B�u�F�X���[�_�E��
	usint		acc2_Act;				//	�A�N�e�B�u�F�����x�Q
	usint		Delay_Act;				//	�A�N�e�B�u�F�����f�B���C
	usint		acc6;					//	���������x�U(0.1)
	usint		accSw_Wf;				//	���������x�֑ؑ�����(0.1)
										//----- AWC -----
	usint		awc_select;				//�w�ߒl����L���I���i�O�F�����^�P�F�L��j
	usint		init_curr_percent;		//���������d���i���j
	usint		init_volt_adj;			//���������d�������l�i�ꌳ�d���ɑ΂��钲���l�F�f�[�^�~0.1V�@���f�[�^�͂Q���݁j
	usint		all_init_time;			//�����������ԁi�����j
	usint		init_up_time;			//���������X�^�[�g�X���[�v���ԁi�����j
	usint		init_dw_time;			//���������_�E���X���[�v���ԁi�����j
	usint		init_limit_curr;		//���������d�������i�f�[�^�~�P�`�@���f�[�^�͂Q���݁j
	usint		crat_curr_percent;		//�N���[�^�����d���i���j
	usint		crat_volt_adj;			//�N���[�^�d�������l�i�ꌳ�d���ɑ΂��钲���l�F�f�[�^�~0.1V�@���f�[�^�͂Q���݁j
	usint		all_crat_time;			//�N���[�^�������ԁi�����j
	usint		crat_dw_time;			//�N���[�^�����_�E���X���[�v���ԁi�����j
	usint		crat_skip_time;			//�N���[�^�X�L�b�v���� (msec)
										//----- NEW_AWC -----
	float		init_up_a;				//���������X�^�[�g�X���[�v�@��
	float		init_up_b;				//���������X�^�[�g�X���[�v�@��
	float		init_up_c;				//���������X�^�[�g�X���[�v�@��
	usint		init_up_min;			//���������X�^�[�g�X���[�v�@������
	usint		init_up_max;			//���������X�^�[�g�X���[�v�@������
} PARMTBL, *PPARMTBL;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		���Œ�p�����[�^�S�`�U�i�f�[�^�錾���ł͕s�v�j							*/
/*			NOTE�F	�u�o1�p�𗬗p�ł���H						�F	2025.03.14	*/
/*				�F	@-1206047_00								�F	----.--.--	*/
/*				�F	���ڐ����T�O�S���ڂɊg��(H379-H504��ǉ�)	�F	2025.04.11	*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	usint				Parm[504];				//	���Œ�p�����[�^�S�`�U		[378]->[504]	//[2025.04.11]
} WELDPARM_4_6, *PWELDPARM_4_6;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�σp�����[�^�̃e�[�u�������o�^�G���A���e�[�u��						*/
/*			NOTE�F	�m�q�P�p�ɐV��								�F	2025.03.14	*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	sint		V_number;					//	�u�ԍ�
	float		Coefficient;				//	�W��
} CALPARMLIST, *PCALPARMLIST;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�σp�����[�^�̃e�[�u�������f�[�^�e�[�u��								*/
/*			NOTE�F	�m�q�P�p�ɐV��								�F	2025.03.14	*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	sint		Data[128];					//	�e�[�u�������f�[�^
} CALPARMDATATBL, *PCALPARMDATATBL;
//
//
//
#define	PRESET_CALPARM_LIST_MAX		16		// �σp�����[�^�̃e�[�u�������f�[�^�e�[�u���̍ő吔�F�P�U�i�G���h�}�[�N����)	23->30	[2025.03.27]	30->16	[2025.04.11]
//
#define	CALC_PARM_MAX_ITEM			160		// �σp�����[�^�ő區�ڐ��F�P�U�O����(�ŏI����[160]�܂�)	+++++[2025.03.27]-----
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�n�ڏ����e�[�u��														*/
/*			NOTE�F	�m�q�P�p									�F	2025.03.14	*/
/*				�F�@350NR1,500NR1�����ŏI�d�l�Ή�				�F	2025.04.11	*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	WELDCODE		WeldCode;										//	�n�ڎ�ʃR�[�h
	A2STBL			A2S_Pulse;										//	���C�������e�[�u���i�p���X)
	S2VTBL			S2V_Pulse;										//	�ꌳ�d���e�[�u���i�p���X�j
	A2STBL			A2S_Short;										//	���C�������e�[�u���i�Z���j
	S2VTBL			S2V_Short;										//	�ꌳ�d���e�[�u���i�Z���j
	WELDPARM		WeldParm;										//	���Œ�p�����[�^
	DCCALPARM		CalParm[116];									//	�σp�����[�^�W���i�u�σp�����[�^�W���F64�v�{�u�σp�����[�^�W���Q�F32�v�{�u�ǉ��σp�����[�^�W���F20�v�j	[2025.04.11]
	PARMTBL			ParmTbl_Pls;									//	�p�����[�^�e�[�u���i�p���X�j
	PARMTBL			ParmTbl_Short;									//	�p�����[�^�e�[�u���i�Z���j
																	//
	CALPARMLIST		CalParmList[PRESET_CALPARM_LIST_MAX];			//	�σp�����[�^�̃e�[�u���������X�g�e�[�u��(�G���h�R�[�h���܂ߍő�P�U)	[32]->[16]	//[2025.04.11]
																	//
	sint			V05_Data[128];									//		�u�T�e�[�u��
	sint			V06_Data[128];									//		�u�U�e�[�u��
	sint			V08_Data[128];									//		�u�W�e�[�u��
	sint			V12_Data[128];									//		�u�P�Q�e�[�u��
	sint			V32_Data[128];									//		�u�R�Q�e�[�u��
	sint			V34_Data[128];									//		�u�R�S�e�[�u��
	sint			V36_Data[128];									//		�u�R�U�e�[�u��
	sint			V56_Data[128];									//		�u�T�U�e�[�u��
	sint			V59_Data[128];									//		�u�T�X�e�[�u��
	sint			V68_Data[128];									//		�u�U�W�e�[�u��
	sint			V13_Data[128];									//		�u�P�R�e�[�u��
	sint			V15_Data[128];									//		�u�P�T�e�[�u��
	sint			V18_Data[128];									//		�u�P�W�e�[�u��
	sint			V19_Data[128];									//		�u�P�X�e�[�u��
	sint			V20_Data[128];									//		�u�Q�O�e�[�u��
	sint			V94_Data[128];									//		�u�X�S�e�[�u��
	sint			V95_Data[128];									//		�u�X�T�e�[�u��
	sint			V57_Data[128];									//		�u�T�V�e�[�u��
	sint			V93_Data[128];									//		�u�X�R�e�[�u��
																	//
	CALPARMDATATBL	CalParmDataTable[PRESET_CALPARM_LIST_MAX-1];	//	�σp�����[�^�̃e�[�u�������f�[�^�e�[�u��(�ő�P�T(�G���h�}�[�N�G���A�͖���))	[32]->[15]	[2025.04.11]
																	//
																	//	�Z���p�n�ڃi�r�f�[�^
	float			Navi_Pram1[7];									//		�s�p����f�[�^
	float			Navi_Pram2[7];									//		�d�ˌp����f�[�^
	float			Navi_Pram3[7];									//		�˂����킹�f�[�^
																	//	�p���X�p�n�ڃi�r�f�[�^
	float			Navi_P_Pram1[7];								//		�s�p����f�[�^
	float			Navi_P_Pram2[7];								//		�d�ˌp����f�[�^
	float			Navi_P_Pram3[7];								//		�˂����킹�f�[�^
																	//
} WELDTABLE_GX3, *PWELDTABLE_GX3;	//	WELDTABLE_NR1, *PWELDTABLE_NR1;		[2025.04.11]
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�n�ڎ�ʖ��̒�`														*/
/*			NOTE�F	�u�o1�p�𗬗p�ł���H						�F	2025.03.14	*/
/*				�F	@-1211075_00 �r�b�g�t�B�[���h				�F	----.--.--	*/
/*------------------------------------------------------------------------------*/
/*	[�f�[�^����]																*/
/*		uchar		Kind					���(d0:�ގ��Ad1:�n�ږ@)			*/
/*		uchar		Code					�R�[�h								*/
/*		char		Name[16]				����								*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	uchar		Kind			:	8;		//	���(d0:�ގ��Ad1:�n�ږ@)
	uchar		Code			:	8;		//	�R�[�h
	char		Name_1			:	8;		//	����
	char		Name_2			:	8;
	char		Name_3			:	8;
	char		Name_4			:	8;
	char		Name_5			:	8;
	char		Name_6			:	8;
	char		Name_7			:	8;
	char		Name_8			:	8;
	char		Name_9			:	8;
	char		Name_10			:	8;
	char		Name_11			:	8;
	char		Name_12			:	8;
	char		Name_13			:	8;
	char		Name_14			:	8;
	char		Name_15			:	8;
	char		Name_16			:	8;
} WELDNAME, *PWELDNAME;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�n�ڃe�[�u���o�[�W����													*/
/*			NOTE�F	�u�o1�p�𗬗p�ł���H						�F	2025.03.14	*/
/*				�F	@-1211075_00								�F	----.--.--	*/
/*------------------------------------------------------------------------------*/
typedef struct
{
 	char		Name_1			:	8;
	char		Name_2			:	8;
	char		Name_3			:	8;
	char		Name_4			:	8;
	char		Name_5			:	8;
	char		Name_6			:	8;
	char		Name_7			:	8;
	char		Name_8			:	8;
	char		Name_9			:	8;
	char		Name_10			:	8;
	char		Name_11			:	8;
	char		Name_12			:	8;
	char		Name_13			:	8;
	char		Name_14			:	8;
	char		Name_15			:	8;
	char		Name_16			:	8;
} WLDTBL_VER, *PWLDTBL_VER;
//
/*------------------------------------------------------------------------------*/
/*	[�T�v]																		*/
/*		�n�ڏ����e�[�u���F�w�b�_�[												*/
/*			NOTE�F	�u�o1�p�𗬗p�ł���H						�F	2025.03.14	*/
/*------------------------------------------------------------------------------*/
typedef	struct
{
	char				Type_1		:	8;		//	���ʎq
	char				Type_2		:	8;		//
	char				Type_3		:	8;		//
	char				Type_4		:	8;		//
	char				Type_5		:	8;		//
	char				Type_6		:	8;		//
	char				Type_7		:	8;		//
	char				Type_8		:	8;		//
	char				Type_9		:	8;		//
	char				Type_10		:	8;		//
	char				Type_11		:	8;		//
	char				Type_12		:	8;		//
	char				Type_13		:	8;		//
	char				Type_14		:	8;		//
	char				Type_15		:	8;		//
	char				Type_16		:	8;		//
	char				*pVersion;				//	�o�[�W����
//	PWLDTBL_VER			pVersion;				//	�o�[�W����
	PWELDNAME			pNamTbl;				//	�n�ڎ�ʖ��̒�`�e�[�u��
//	PWELDTABLE_NR1		pWldTbl;				//	�n�ڏ����e�[�u��
	PWELDTABLE_GX3		pWldTbl;				//	�n�ڏ����e�[�u��	[2025.04.11]
} H_WELDTABLE_GX3, *PH_WELDTABLE_GX3;	//	H_WELDTABLE_NR1, *PH_WELDTABLE_NR1;		[2025.04.11]
/*------------------------------------------------------------------------------*/
/* end of file 																	*/
/*------------------------------------------------------------------------------*/
